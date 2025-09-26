package dto

import (
	"context"
	"github.com/google/uuid"
)

type Message struct {
	Ctx     context.Context
	EventId uuid.UUID `json:"event_id"`
	Body    TaskEvent `json:"body"`
}
