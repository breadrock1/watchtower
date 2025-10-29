package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"watchtower/internal/domain/core/structures"
)

type Task struct {
	ID         uuid.UUID
	Bucket     string
	FilePath   string
	FileSize   int64
	CreatedAt  time.Time
	ModifiedAt time.Time
	Status     int
	StatusText string
}

func (te *Task) ToDomain() *domain.Task {
	return &domain.Task{
		ID:         te.ID,
		Bucket:     te.Bucket,
		FilePath:   te.FilePath,
		FileSize:   te.FileSize,
		StatusText: te.StatusText,
		Status:     domain.TaskStatus(te.Status),
		CreatedAt:  te.CreatedAt,
		ModifiedAt: te.ModifiedAt,
	}
}

func FromDomainTask(taskEvent *domain.Task) Task {
	return Task{
		taskEvent.ID,
		taskEvent.Bucket,
		taskEvent.FilePath,
		taskEvent.FileSize,
		taskEvent.CreatedAt,
		taskEvent.ModifiedAt,
		int(taskEvent.Status),
		taskEvent.StatusText,
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
