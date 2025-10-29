package rmq

import (
	"context"

	"github.com/google/uuid"
	"watchtower/internal/application/models"
)

type Message struct {
	Ctx     context.Context
	EventId uuid.UUID   `json:"event_id"`
	Body    models.Task `json:"body"`
}

func (m *Message) ToMessage() *models.Message {
	return &models.Message{
		Ctx:     m.Ctx,
		EventId: m.EventId,
		Body:    m.Body,
	}
}
