package redis

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"watchtower/internal/application/models"
)

type RedisValue struct {
	ID         string `json:"id"`
	Bucket     string `json:"bucket"`
	FilePath   string `json:"file_path"`
	FileSize   int64  `json:"file_size"`
	CreatedAt  int64  `json:"created_at"`
	ModifiedAt int64  `json:"modified_at"`
	Status     int    `json:"status"`
	StatusText string `json:"status_text"`
	EventType  int    `json:"event_type"`
}

func (rv *RedisValue) ConvertToTaskEvent() (*models.TaskEvent, error) {
	modDt := time.Unix(rv.ModifiedAt, 0)
	createDt := time.Unix(rv.CreatedAt, 0)
	taskID, err := uuid.FromBytes([]byte(rv.ID))
	if err != nil {
		return nil, fmt.Errorf("invalid task id: %w", err)
	}

	event := &models.TaskEvent{
		ID:         taskID,
		Bucket:     rv.Bucket,
		FilePath:   rv.FilePath,
		FileSize:   rv.FileSize,
		CreatedAt:  createDt,
		ModifiedAt: modDt,
		Status:     rv.Status,
		StatusText: rv.StatusText,
	}

	return event, nil
}

func ConvertFromTaskEvent(taskEvent *models.TaskEvent) *RedisValue {
	return &RedisValue{
		ID:         taskEvent.ID.String(),
		Bucket:     taskEvent.Bucket,
		FilePath:   taskEvent.FilePath,
		FileSize:   taskEvent.FileSize,
		CreatedAt:  taskEvent.CreatedAt.Unix(),
		ModifiedAt: taskEvent.ModifiedAt.Unix(),
		Status:     taskEvent.Status,
		StatusText: taskEvent.StatusText,
	}
}
