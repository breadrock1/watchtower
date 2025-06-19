package task_manager

import (
	"context"

	"watchtower/internal/application/dto"
)

type ITaskManager interface {
	Push(ctx context.Context, task *dto.TaskEvent) error
	GetAll(ctx context.Context, bucket string) ([]*dto.TaskEvent, error)
	Get(ctx context.Context, bucket, file string) (*dto.TaskEvent, error)
}
