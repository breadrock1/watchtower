package redis

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"watchtower/internal/application/dto"
)

type RedisValue struct {
	Id         string `json:"id"`
	Bucket     string `json:"bucket"`
	FilePath   string `json:"file_path"`
	FileSize   int64  `json:"file_size"`
	CreatedAt  int64  `json:"created_at"`
	ModifiedAt int64  `json:"modified_at"`
	Status     int    `json:"status"`
	StatusText string `json:"status_text"`
}

func (rv *RedisValue) ConvertToTaskEvent() (*dto.TaskEvent, error) {
	id, err := uuid.Parse(rv.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize task id: %s", rv.Id)
	}

	status := dto.TaskStatus(rv.Status)
	modDt := time.Unix(rv.ModifiedAt, 0)
	createDt := time.Unix(rv.CreatedAt, 0)

	event := &dto.TaskEvent{
		Id:         id,
		Bucket:     rv.Bucket,
		FilePath:   rv.FilePath,
		FileSize:   rv.FileSize,
		CreatedAt:  createDt,
		ModifiedAt: modDt,
		Status:     status,
		StatusText: rv.StatusText,
	}

	return event, nil
}

func ConvertFromTaskEvent(te *dto.TaskEvent) *RedisValue {
	return &RedisValue{
		Id:         te.Id.String(),
		Bucket:     te.Bucket,
		FilePath:   te.FilePath,
		FileSize:   te.FileSize,
		CreatedAt:  te.CreatedAt.Unix(),
		ModifiedAt: te.ModifiedAt.Unix(),
		Status:     te.Status.TaskStatusToInt(),
	}
}
