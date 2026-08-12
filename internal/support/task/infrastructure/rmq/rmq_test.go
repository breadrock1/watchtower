package rmq

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
	"testing"
)

const (
	testTraceIDHex = "00112233445566778899aabbccddeeff"
	testSpanIDHex  = "0011223344556677"
)

func TestRabbitMQClient(t *testing.T) {
	t.Run("Extract context from nil headers", func(t *testing.T) {
		ctx := extractSpanContextFromHeaders(nil)

		assert.NotNil(t, ctx, "expected context")
	})

	t.Run("Inject and extract span context headers", func(t *testing.T) {
		traceID, _ := trace.TraceIDFromHex(testTraceIDHex)
		spanID, _ := trace.SpanIDFromHex(testSpanIDHex)
		ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: traceID,
			SpanID:  spanID,
			Remote:  true,
		}))

		headers := injectSpanContextToHeaders(ctx)
		extracted := extractSpanContextFromHeaders(amqp.Table{
			"trace-id":    headers["trace-id"],
			"span-id":     headers["span-id"],
			"trace-flags": headers["trace-flags"],
			"trace-state": headers["trace-state"],
		})

		assert.Equal(t, traceID, trace.SpanContextFromContext(extracted).TraceID(), "unexpected extracted trace id")
	})
}
