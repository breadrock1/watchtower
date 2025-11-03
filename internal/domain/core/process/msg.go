package process

import (
	"context"

	"github.com/google/uuid"
)

type MessageID = uuid.UUID

type Message struct {
	Ctx     context.Context
	EventId MessageID
	Body    Task
}
