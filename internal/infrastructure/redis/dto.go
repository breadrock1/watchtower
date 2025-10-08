package redis

import (
	"time"

	"watchtower/internal/application/dto"
	"watchtower/internal/application/mapping"
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

func (rv *RedisValue) ConvertToTaskEvent() (*dto.TaskEvent, error) {
	modDt := time.Unix(rv.ModifiedAt, 0)
	createDt := time.Unix(rv.CreatedAt, 0)

	event := &dto.TaskEvent{
		ID:         rv.ID,
		Bucket:     rv.Bucket,
		FilePath:   rv.FilePath,
		FileSize:   rv.FileSize,
		CreatedAt:  createDt,
		ModifiedAt: modDt,
		Status:     mapping.TaskStatusFromInt(rv.Status),
		StatusText: rv.StatusText,
	}

	return event, nil
}

func ConvertFromTaskEvent(taskEvent *dto.TaskEvent) *RedisValue {
	return &RedisValue{
		ID:         taskEvent.ID,
		Bucket:     taskEvent.Bucket,
		FilePath:   taskEvent.FilePath,
		FileSize:   taskEvent.FileSize,
		CreatedAt:  taskEvent.CreatedAt.Unix(),
		ModifiedAt: taskEvent.ModifiedAt.Unix(),
		Status:     int(taskEvent.Status),
		StatusText: taskEvent.StatusText,
	}
}
