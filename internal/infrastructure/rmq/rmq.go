package rmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"watchtower/internal/infrastructure/httpserver"

	amqp "github.com/rabbitmq/amqp091-go"
	"watchtower/internal/application/dto"
)

const ConsumerName = "watchtower-consumer"

type RmqClient struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	config   *Config
	redirect chan dto.Message
	done     chan error
}

func New(config *Config) (*RmqClient, error) {
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

	client := &RmqClient{
		conn,
		rmqCh,
		config,
		make(chan dto.Message),
		make(chan error),
	}

	return client, nil
}

func (r *RmqClient) GetConsumerChannel() chan dto.Message {
	return r.redirect
}

func (r *RmqClient) Publish(ctx context.Context, msg dto.Message) error {
	ctx, span := httpserver.GlobalTracer.Start(ctx, "rmq-publish")
	defer span.End()

	headers := injectSpanContextToHeaders(ctx)
	body, err := json.Marshal(msg)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed while marshalling rmq body: %w", err)
	}

	err = r.channel.Publish(
		r.config.Exchange,
		r.config.RoutingKey,
		false,
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
		return fmt.Errorf("failed while publishing rmq message: %w", err)
	}

	span.SetStatus(codes.Ok, "success")
	return nil
}

func (r *RmqClient) Consume(_ context.Context) error {
	go r.handleReconnect()

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
		return fmt.Errorf("failed to init rmq consumer: %w", err)
	}

	go r.handle(deliveries, r.done)

	return nil
}

func (r *RmqClient) StopConsuming(_ context.Context) error {
	if err := r.channel.Cancel(ConsumerName, true); err != nil {
		return fmt.Errorf("consumer cancel failed: %w", err)
	}

	if err := r.conn.Close(); err != nil {
		return fmt.Errorf("amqp connection close error: %w", err)
	}

	// wait for handle() to exit
	return <-r.done
}

func (r *RmqClient) handle(deliveries <-chan amqp.Delivery, done chan error) {
	cleanup := func() {
		log.Printf("handle: deliveries channel closed")
		done <- nil
	}

	defer cleanup()

	for delMsg := range deliveries {
		ctx := extractSpanContextFromHeaders(delMsg.Headers)
		span := trace.SpanFromContext(ctx)
		defer span.End()

		msg := &dto.Message{}
		err := json.Unmarshal(delMsg.Body, msg)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			slog.Error("failed while read rmq message", slog.String("err", err.Error()))
			continue
		}

		span.SetName("rmq-consume")
		span.SetAttributes(attribute.String("task-id", msg.EventId.String()))

		msg.Ctx = ctx
		r.redirect <- *msg
	}
}

func (r *RmqClient) handleReconnect() {
	for {
		select {
		case <-r.done:
			return

		case <-r.conn.NotifyClose(make(chan *amqp.Error)):
			log.Println("Attempting to reconnect...")

			rmqConfig := amqp.Config{
				Properties: amqp.NewConnectionProperties(),
				Heartbeat:  10 * time.Second,
			}

			rmqConfig.Properties.SetClientConnectionName(ConsumerName)

			var err error
			for reconnCounter := 0; reconnCounter < 5; reconnCounter++ {
				r.conn, err = amqp.DialConfig(r.config.Address, rmqConfig)
				if err != nil {
					log.Printf("failed while re-connecting to rmq: %v", err)
					return
				}

				r.channel, err = r.conn.Channel()
				if err == nil {
					log.Printf("connection to rmq has been returned!")
					break
				}

				log.Printf("failed to create rmq channel: %v", err)
				time.Sleep(time.Duration(reconnCounter*reconnCounter) * time.Second)
			}

			if err != nil {
				log.Println("Failed to reconnect to RabbitMQ")
				return
			}
		}
	}
}

func (r *RmqClient) CreateExchange(name string) error {
	err := r.channel.ExchangeDeclare(name, "direct", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed while declaring exchange: %w", err)
	}

	return nil
}

func (r *RmqClient) CreateQueue(exchange, queue, routingKey string) error {
	_, err := r.channel.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed while declaring queue: %w", err)
	}

	err = r.channel.QueueBind(queue, routingKey, exchange, false, nil)
	if err != nil {
		return fmt.Errorf("failed while binding queue: %w", err)
	}

	return nil
}

func injectSpanContextToHeaders(ctx context.Context) amqp.Table {
	carrier := propagation.HeaderCarrier{}
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, carrier)

	span := trace.SpanFromContext(ctx)
	sCtx := span.SpanContext()

	headers := amqp.Table{}
	headers["trace-id"] = sCtx.TraceID().String()
	headers["span-id"] = sCtx.SpanID().String()
	headers["trace-flags"] = sCtx.TraceFlags().String()
	headers["trace-state"] = sCtx.TraceState().String()

	return headers
}

func extractSpanContextFromHeaders(headers amqp.Table) context.Context {
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
