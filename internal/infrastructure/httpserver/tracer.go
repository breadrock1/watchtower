package httpserver

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func InitTracer(config *Config) (trace.Tracer, error) {
	tracerConfig := config.Tracer

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.TelemetrySDKLanguageGo,
		semconv.ServiceNameKey.String(AppName),
	)

	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(tracerConfig.Address),
		otlptracegrpc.WithInsecure(),
	)

	ctx := context.Background()
	traceExporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, err
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	sampler := sdktrace.AlwaysSample()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	otel.SetTracerProvider(tp)

	tracer = tp.Tracer(AppName)

	return tracer, nil
}

// GetTracer gets global Tracer
// func GetTracer() trace.Tracer {
func GetTracer() trace.Tracer {
	return tracer
}

//func tracingFilter() echo.MiddlewareFunc {
//	return otelecho.Middleware(AppName,
//		otelecho.WithTracerProvider(otel.GetTracerProvider()),
//		otelecho.WithPropagators(otel.GetTextMapPropagator()),
//		otelecho.WithSkipper(func(c echo.Context) bool {
//			return shouldSkipTrace(c.Path())
//		}),
//	)
//}
//
//func shouldSkipTrace(path string) bool {
//	// List of paths to exclude from tracing
//	excludedPaths := []string{
//		"/health",
//		"/metrics",
//		"/favicon.ico",
//		"/static/",
//	}
//
//	for _, excluded := range excludedPaths {
//		if strings.HasPrefix(path, excluded) {
//			return true
//		}
//	}
//	return false
//}
