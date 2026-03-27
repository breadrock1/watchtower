package domain

import (
	"watchtower/internal/shared/kernel"
)

// ITaskQueue defines the complete interface for task queue operations.
// It combines publishing and consuming capabilities for a complete
// producer-consumer pattern implementation.
type ITaskQueue interface {
	IConsumer
	IPublisher
}

// IPublisher defines operations for publishing tasks to the queue.
// Publishers are responsible for enqueuing tasks for asynchronous processing.
type IPublisher interface {
	// Publish sends a task message to the queue for asynchronous processing.
	// The message is persisted in the queue and will be delivered to a consumer.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - msg: Complete message containing the task and metadata
	//
	// Returns:
	//   - error: ErrQueueUnavailable if queue service is down,
	//            ErrInvalidMessage if message validation fails,
	//            ErrPublishTimeout if operation exceeds deadline,
	//            or other queue-specific errors
	//
	// Example:
	//   task := Task{
	//       ID:             uuid.New(),
	//       BucketID:       "input-bucket",
	//       ObjectID:       "data/file.json",
	//       Status:         Received,
	//       CreatedAt:      time.Now(),
	//   }
	//
	//   msg := Message{
	//       EventId: uuid.New(),
	//       Body:    task,
	//       Metadata: map[string]string{"source": "api"},
	//   }
	//
	//   err := publisher.Publish(ctx, msg)
	//   if err != nil {
	//       log.Printf("Failed to publish task: %v", err)
	//   }
	Publish(ctx kernel.Ctx, msg Message) error
}

// IConsumer defines operations for consuming tasks from the queue.
// Consumers process tasks asynchronously and manage the consumption lifecycle.
type IConsumer interface {
	// GetConsumerChannel returns a read-only channel for receiving messages.
	// This channel should be used in a select statement or range loop to
	// process incoming tasks.
	//
	// Returns:
	//   - chan Message: Channel that delivers messages as they arrive
	//
	// Example:
	//   msgChan := consumer.GetConsumerChannel()
	//   for msg := range msgChan {
	//       go processMessage(msg)
	//   }
	GetConsumerChannel() chan Message

	// StartConsuming begins the message consumption process.
	// This method typically connects to the queue service and begins
	// delivering messages to the consumer channel.
	//
	// Parameters:
	//   - ctx: Context for controlling the consumption lifecycle
	//
	// Returns:
	//   - error: ErrConsumerAlreadyStarted if already consuming,
	//            ErrQueueUnavailable if cannot connect to queue,
	//            or other queue-specific errors
	//
	// Example:
	//   ctx, cancel := context.WithCancel(context.Background())
	//   defer cancel()
	//
	//   go func() {
	//       if err := consumer.StartConsuming(ctx); err != nil {
	//           log.Printf("Consumer failed: %v", err)
	//       }
	//   }()
	StartConsuming(ctx kernel.Ctx) error

	// StopConsuming gracefully stops the message consumption process.
	// It should complete any in-progress message handling before stopping.
	//
	// Parameters:
	//   - ctx: Context for timeout control during shutdown
	//
	// Returns:
	//   - error: ErrConsumerNotStarted if not consuming,
	//            ErrShutdownTimeout if graceful shutdown times out,
	//            or other queue-specific errors
	//
	// Example:
	//   // Graceful shutdown
	//   shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	//   defer cancel()
	//
	//   if err := consumer.StopConsuming(shutdownCtx); err != nil {
	//       log.Printf("Force shutdown: %v", err)
	//   }
	StopConsuming(ctx kernel.Ctx) error
}
