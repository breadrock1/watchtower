package redis

import (
	"fmt"
	"time"
	"watchtower/internal/domain/core/process"

	"github.com/google/uuid"
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

func (rv *RedisValue) ConvertToTask() (*process.Task, error) {
	taskID, err := uuid.Parse(rv.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid task id: %w", err)
	}

	modifiedAt := time.Unix(rv.ModifiedAt, 0)
	createdAt := time.Unix(rv.CreatedAt, 0)

	event := &process.Task{
		ID:         taskID,
		CreatedAt:  createdAt,
		ModifiedAt: modifiedAt,
		BucketID:   rv.Bucket,
		ObjectID:   rv.FilePath,
		StatusText: rv.StatusText,
		Status:     process.TaskStatus(rv.Status),
	}

	return event, nil
}

func ConvertFromTaskEvent(task *process.Task) *RedisValue {
	return &RedisValue{
		ID:         task.ID.String(),
		Bucket:     task.BucketID,
		FilePath:   task.ObjectID,
		CreatedAt:  task.CreatedAt.Unix(),
		ModifiedAt: task.ModifiedAt.Unix(),
		StatusText: task.StatusText,
		Status:     int(task.Status),
	}
}
