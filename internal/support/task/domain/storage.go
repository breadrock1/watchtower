package domain

import (
	"context"

	"watchtower/internal/core/cloud/domain"
)

type ITaskStorage interface {
	ITaskManager
}

type ITaskManager interface {
	GetTask(ctx context.Context, bucketID domain.BucketID, taskID TaskID) (*Task, error)
	GetAllBucketTasks(ctx context.Context, bucketID domain.BucketID) ([]*Task, error)
	UpdateTask(ctx context.Context, task *Task) error
}
