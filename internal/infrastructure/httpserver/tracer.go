package httpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

var (
	excludedPaths = []string{
		"/health",
		"/metrics",
		"/favicon.ico",
		"/static/",
	}
	propagator = propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
	)
)

var GlobalTracer trace.Tracer

func InitTracer(config *Config) (trace.Tracer, error) {
	tracerConfig := config.Tracer
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(tracerConfig.Address),
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

	otel.SetTextMapPropagator(propagator)
	GlobalTracer = tp.Tracer(AppName)
	otel.SetTracerProvider(tp)
	return GlobalTracer, nil
}

func TracerSkipper(c echo.Context) bool {
	for _, excluded := range excludedPaths {
		if strings.HasPrefix(c.Path(), excluded) {
			return true
		}
	}

	if c.Request().Method == "OPTIONS" {
		return true
	}

	return false
}
