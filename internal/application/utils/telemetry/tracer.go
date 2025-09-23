package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	"watchtower/internal/application/services/server"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

const AppName = server.AppName

var (
	GlobalTracer    trace.Tracer
	TracePropagator = propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
)

func InitTracer(config TracerConfig) (trace.Tracer, error) {
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(config.Address),
		otlptracegrpc.WithInsecure(),
	)

	ctx := context.Background()
	traceExporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to otlp trace server: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.TelemetrySDKLanguageGo,
			semconv.ServiceNameKey.String(AppName),
		),
	)
	sampler := sdktrace.AlwaysSample()
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTextMapPropagator(TracePropagator)
	GlobalTracer = tp.Tracer(AppName)
	otel.SetTracerProvider(tp)
	return GlobalTracer, nil
}
