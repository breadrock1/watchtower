package models

import (
	"time"

	"github.com/google/uuid"
	"watchtower/internal/domain/core/structures"
)

type TaskEvent struct {
	ID         uuid.UUID
	Bucket     string
	FilePath   string
	FileSize   int64
	CreatedAt  time.Time
	ModifiedAt time.Time
	Status     int
	StatusText string
}

func (te *TaskEvent) ToDomain() *domain.TaskEvent {
	return &domain.TaskEvent{
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

func FromDomain(taskEvent *domain.TaskEvent) TaskEvent {
	return TaskEvent{
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
