package mapping

import (
	"github.com/google/uuid"

	"watchtower/internal/support/task/domain"
)

func MessageFromTask(task *domain.Task) domain.Message {
	return domain.Message{
		EventId: uuid.New(),
		Body:    *task,
	}
}
