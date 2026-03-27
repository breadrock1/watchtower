package domain

import (
	"watchtower/internal/shared/kernel"
)

// Message represents a task wrapped for queue transport.
// It includes context for distributed tracing and the actual task payload.
type Message struct {
	// Ctx carries cancellation signals and deadlines across service boundaries
	Ctx kernel.Ctx

	// EventId uniquely identifies this specific message in the queue
	EventId kernel.MessageID

	// Body contains the actual task to be processed
	Body Task

	// Metadata holds additional routing and tracing information
	Metadata map[string]string

	// DeliveryAttempt counts how many times this message has been delivered
	DeliveryAttempt int
}
