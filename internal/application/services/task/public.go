package task

import (
	"context"

	"watchtower/internal/application/models"
)

type ITaskQueue interface {
	IConsumer
	IPublisher
}

type ITaskStorage interface {
	ITaskManager
}

type IPublisher interface {
	Publish(ctx context.Context, msg models.Message) error
}

type IConsumer interface {
	Consume(ctx context.Context) error
	GetConsumerChannel() chan models.Message
	StopConsuming(ctx context.Context) error
}

type ITaskManager interface {
	Push(ctx context.Context, task *models.TaskEvent) error
	GetAll(ctx context.Context, bucket string) ([]*models.TaskEvent, error)
	Get(ctx context.Context, bucket, file string) (*models.TaskEvent, error)
}
