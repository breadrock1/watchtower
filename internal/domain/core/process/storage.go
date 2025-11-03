package process

import (
	"context"

	"watchtower/internal/domain/core/cloud"
)

type ITaskStorage interface {
	ITaskManager
}

type ITaskManager interface {
	GetTask(ctx context.Context, bucketID cloud.BucketID, taskID TaskID) (*Task, error)
	GetAllBucketTasks(ctx context.Context, bucketID cloud.BucketID) ([]*Task, error)
	UpdateTask(ctx context.Context, task *Task) error
}
