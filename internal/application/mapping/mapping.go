package mapping

import (
	"fmt"

	"github.com/google/uuid"
	"watchtower/internal/application/models"
	"watchtower/internal/domain/core/structures"
)

func MessageFromTaskEvent(event *domain.TaskEvent) models.Message {
	taskEventDto := models.FromDomain(event)
	return models.Message{
		EventId: uuid.New(),
		Body:    taskEventDto,
	}
}

func TaskStatusFromString(enumVal string) (domain.TaskStatus, error) {
	switch enumVal {
	case "received":
		return domain.Received, nil
	case "pending":
		return domain.Pending, nil
	case "processing":
		return domain.Processing, nil
	case "successful":
		return domain.Successful, nil
	case "failed":
		return domain.Failed, nil
	default:
		return domain.Pending, fmt.Errorf("unknown task status: %s", enumVal)
	}
}

func TaskStatusFromInt(enum int) domain.TaskStatus {
	return domain.TaskStatus(enum)
}
