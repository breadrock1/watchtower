package mapping

import (
	"fmt"

	"github.com/google/uuid"
	"watchtower/internal/application/dto"
)

func MessageFromTaskEvent(event dto.TaskEvent) dto.Message {
	return dto.Message{
		EventId: uuid.New(),
		Body:    event,
	}
}

func TaskStatusFromString(enumVal string) (dto.TaskStatus, error) {
	switch enumVal {
	case "received":
		return dto.Received, nil
	case "pending":
		return dto.Pending, nil
	case "processing":
		return dto.Processing, nil
	case "successful":
		return dto.Successful, nil
	case "failed":
		return dto.Failed, nil
	default:
		return dto.Pending, fmt.Errorf("unknown task status: %s", enumVal)
	}
}
