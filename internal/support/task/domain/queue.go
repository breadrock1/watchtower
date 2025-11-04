package domain

import (
	"context"
)

type ITaskQueue interface {
	IConsumer
	IPublisher
}

type IPublisher interface {
	Publish(ctx context.Context, msg Message) error
}

type IConsumer interface {
	GetConsumerChannel() chan Message
	StartConsuming(ctx context.Context) error
	StopConsuming(ctx context.Context) error
}
