package redis

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"watchtower/internal/application/dto"
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
}

func (rv *RedisValue) ConvertToTaskEvent() (*dto.TaskEvent, error) {
	taskID, err := uuid.Parse(rv.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize task id: %s", rv.ID)
	}

	modDt := time.Unix(rv.ModifiedAt, 0)
	createDt := time.Unix(rv.CreatedAt, 0)

	event := &dto.TaskEvent{
		Id:         taskID,
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

func ConvertFromTaskEvent(taskEvent *dto.TaskEvent) *RedisValue {
	return &RedisValue{
		ID:         taskEvent.Id.String(),
		Bucket:     taskEvent.Bucket,
		FilePath:   taskEvent.FilePath,
		FileSize:   taskEvent.FileSize,
		CreatedAt:  taskEvent.CreatedAt.Unix(),
		ModifiedAt: taskEvent.ModifiedAt.Unix(),
		Status:     taskEvent.Status,
	}
}
