package task_queue

import (
	"context"

	"watchtower/internal/application/dto"
)

type ITaskQueue interface {
	IConsumer
	IPublisher
}

type IPublisher interface {
	Publish(ctx context.Context, msg dto.Message) error
}

type IConsumer interface {
	Consume(ctx context.Context) error
	StopConsuming(ctx context.Context) error
}
