package models

import (
	"context"

	"github.com/google/uuid"
	"watchtower/internal/domain/core/structures"
)

type Message struct {
	Ctx     context.Context
	EventId uuid.UUID
	Body    Task
}

func MessageFromTask(event *domain.Task) Message {
	taskEventDto := FromDomainTask(event)
	return Message{
		EventId: uuid.New(),
		Body:    taskEventDto,
	}
}
