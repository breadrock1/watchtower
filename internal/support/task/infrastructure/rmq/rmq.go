package rmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/breadrock1/otlp-go/otlp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"watchtower/internal/shared/kernel"
	"watchtower/internal/support/task/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

const ConsumerName = "watchtower-consumer"

type RabbitMQClient struct {
	redirect chan domain.Message
	config   Config

	conn    *amqp.Connection
	channel *amqp.Channel
}

func New(config Config) (domain.ITaskQueue, error) {
	var rmqClient RabbitMQClient
	rmqConfig := amqp.Config{
		Properties: amqp.NewConnectionProperties(),
		Heartbeat:  10 * time.Second,
	}
	rmqConfig.Properties.SetClientConnectionName(ConsumerName)

	conn, err := amqp.DialConfig(config.Address, rmqConfig)
	if err != nil {
		return nil, fmt.Errorf("failed while connecting to rmq: %w", err)
	}

	rmqCh, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create rmq channel: %w", err)
	}

	slog.Info("rmq connection established", slog.String("address", config.Address))

	rmqClient = RabbitMQClient{
		make(chan domain.Message),
		config,
		conn,
		rmqCh,
	}

	return &rmqClient, nil
}

func (r *RabbitMQClient) GetConsumerChannel() chan domain.Message {
	return r.redirect
}

func (r *RabbitMQClient) Publish(ctx kernel.Ctx, msg domain.Message) error {
	ctx, span := otlp_go.GlobalTracer.Start(ctx, "rmq-publish")
	defer span.End()

	headers := injectSpanContextToHeaders(ctx)
	body, err := json.Marshal(msg)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("rmq: serialization error: %w", err)
	}

	err = r.channel.Publish(
		r.config.Exchange,
		r.config.RoutingKey,
		true,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Headers:     headers,
			Body:        body,
			Timestamp:   time.Now(),
		},
	)

	span.SetAttributes(
		attribute.Int("message.size", len(body)),
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.destination", r.config.Exchange),
		attribute.String("messaging.rabbitmq.routing_key", r.config.RoutingKey),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("rmq: publish error: %w", err)
	}

	span.SetStatus(codes.Ok, "success")
	return nil
}

func (r *RabbitMQClient) StartConsuming(ctx kernel.Ctx) error {
	go r.handleReconnect(ctx)

	deliveries, err := r.channel.Consume(
		r.config.QueueName, // name
		ConsumerName,       // consumerTag,
		true,               // autoAck
		false,              // exclusive
		false,              // noLocal
		false,              // noWait
		nil,                // arguments
	)

	if err != nil {
		return fmt.Errorf("rmq: consume error: %w", err)
	}

	go r.handleMessage(ctx, deliveries)

	return nil
}

func (r *RabbitMQClient) StopConsuming(_ kernel.Ctx) error {
	if err := r.channel.Cancel(ConsumerName, true); err != nil {
		return fmt.Errorf("rmq: consumer cancel failed: %w", err)
	}

	if err := r.conn.Close(); err != nil {
		return fmt.Errorf("rmq: close connection failed: %w", err)
	}

	return nil
}

func (r *RabbitMQClient) handleMessage(ctx kernel.Ctx, deliveries <-chan amqp.Delivery) {
	slog.Info("launching rmq consumer")

	for {
		select {
		case <-ctx.Done():
			if err := r.StopConsuming(ctx); err != nil {
				slog.Error("failed to stop rmq consuming", slog.String("err", err.Error()))
				return
			}

			slog.Info("rmq deliveries channel has been closed")

			return

		case delMsg, ok := <-deliveries:
			if !ok {
				continue
			}

			spanCtx := extractSpanContextFromHeaders(delMsg.Headers)
			span := trace.SpanFromContext(spanCtx)

			consumeMsg := &Message{}
			err := json.Unmarshal(delMsg.Body, consumeMsg)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				slog.Error("rmq: failed while deserialize msg", slog.String("err", err.Error()))
				continue
			}

			span.SetName("rmq-consume")
			span.SetAttributes(attribute.String("task-id", consumeMsg.EventId.String()))

			consumeMsg.Ctx = spanCtx
			msg := consumeMsg.ToMessage()

			r.redirect <- *msg
			span.End()
		}
	}
}

func (r *RabbitMQClient) handleReconnect(ctx kernel.Ctx) {
	for {
		select {
		case <-ctx.Done():
			return

		case <-r.conn.NotifyClose(make(chan *amqp.Error)):
			slog.Warn("attempting to reconnect...")

			rmqConfig := amqp.Config{
				Properties: amqp.NewConnectionProperties(),
				Heartbeat:  10 * time.Second,
			}

			rmqConfig.Properties.SetClientConnectionName(ConsumerName)

			var err error
			var reconnectDelay int
			for reconnectCounter := 0; reconnectCounter < 5; reconnectCounter++ {
				r.conn, err = amqp.DialConfig(r.config.Address, rmqConfig)
				if err != nil {
					slog.Warn("rmq: failed while re-connecting", slog.String("err", err.Error()))
					return
				}

				r.channel, err = r.conn.Channel()
				if err == nil {
					slog.Warn("rmq: connection has been returned")
					break
				}

				slog.Error("rmq: failed to create channel", slog.String("err", err.Error()))
				reconnectDelay = reconnectCounter * reconnectCounter
				time.Sleep(time.Duration(reconnectDelay) * time.Second)
			}

			if err != nil {
				slog.Error("rmq: failed to restore connection", slog.String("err", err.Error()))
				return
			}
		}
	}
}

func injectSpanContextToHeaders(ctx kernel.Ctx) amqp.Table {
	carrier := propagation.HeaderCarrier{}
	otlp_go.TracePropagator.Inject(ctx, carrier)

	span := trace.SpanFromContext(ctx)
	sCtx := span.SpanContext()

	headers := amqp.Table{}
	headers["trace-id"] = sCtx.TraceID().String()
	headers["span-id"] = sCtx.SpanID().String()
	headers["trace-flags"] = sCtx.TraceFlags().String()
	headers["trace-state"] = sCtx.TraceState().String()

	return headers
}

func extractSpanContextFromHeaders(headers amqp.Table) kernel.Ctx {
	ctx := context.Background()
	if headers == nil {
		return ctx
	}

	traceID, _ := trace.TraceIDFromHex(headers["trace-id"].(string))
	spanID, _ := trace.SpanIDFromHex(headers["span-id"].(string))
	sCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
		TraceState: trace.TraceState{},
		Remote:     true,
	})

	return trace.ContextWithSpanContext(ctx, sCtx)
}
