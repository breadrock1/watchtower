package rmq

import (
	"context"

	"github.com/google/uuid"
	"watchtower/internal/support/task/domain"
)

type Message struct {
	Ctx     context.Context
	EventId uuid.UUID   `json:"event_id"`
	Body    domain.Task `json:"body"`
}

func (m *Message) ToMessage() *domain.Message {
	return &domain.Message{
		Ctx:     m.Ctx,
		EventId: m.EventId,
		Body:    m.Body,
	}
}
