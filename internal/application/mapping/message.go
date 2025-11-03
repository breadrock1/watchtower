package mapping

import (
	"github.com/google/uuid"

	"watchtower/internal/domain/core/process"
)

func MessageFromTask(task *process.Task) process.Message {
	return process.Message{
		EventId: uuid.New(),
		Body:    *task,
	}
}
