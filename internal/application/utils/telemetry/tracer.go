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
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
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
	sampler := sdktrace.AlwaysSample()
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.TelemetrySDKLanguageGo,
			semconv.ServiceNameKey.String(AppName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to merge trace resources: %w", err)
	}

	var traceOpts []sdktrace.TracerProviderOption
	traceOpts = append(traceOpts, sdktrace.WithResource(res))
	traceOpts = append(traceOpts, sdktrace.WithSampler(sampler))

	if config.EnableJaeger {
		ctx := context.Background()
		client := otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(config.Address),
			otlptracegrpc.WithInsecure(),
		)

		traceExporter, err := otlptrace.New(ctx, client)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to otlp trace server: %w", err)
		}
		bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
		traceOpts = append(traceOpts, sdktrace.WithSpanProcessor(bsp))
	}

	tp := sdktrace.NewTracerProvider(traceOpts...)
	otel.SetTextMapPropagator(TracePropagator)
	GlobalTracer = tp.Tracer(AppName)
	otel.SetTracerProvider(tp)
	return GlobalTracer, nil
}
