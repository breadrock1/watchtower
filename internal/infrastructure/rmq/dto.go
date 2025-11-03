package rmq

import (
	"context"

	"github.com/google/uuid"
	"watchtower/internal/domain/core/process"
)

type Message struct {
	Ctx     context.Context
	EventId uuid.UUID    `json:"event_id"`
	Body    process.Task `json:"body"`
}

func (m *Message) ToMessage() *process.Message {
	return &process.Message{
		Ctx:     m.Ctx,
		EventId: m.EventId,
		Body:    m.Body,
	}
}
