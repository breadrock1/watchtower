package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"watchtower/internal/application/mapping"
	"watchtower/internal/application/models"
	"watchtower/internal/application/services/task"
	"watchtower/internal/application/utils/telemetry"
	"watchtower/internal/domain/core/structures"
)

type TaskUseCase struct {
	storage task.ITaskManager
	queue   task.ITaskQueue
}

func NewTaskUseCase(storage task.ITaskManager, queue task.ITaskQueue) *TaskUseCase {
	return &TaskUseCase{storage, queue}
}

func (t *TaskUseCase) GetConsumeChannel() chan models.Message {
	return t.queue.GetConsumerChannel()
}

func (t *TaskUseCase) GetAllTasks(ctx context.Context, bucket string) ([]*models.TaskEvent, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "load-all-tasks")
	defer span.End()

	span.SetAttributes(attribute.String("bucket", bucket))

	tasks, err := t.storage.GetAll(ctx, bucket)
	if err != nil {
		err = fmt.Errorf("task manager error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return tasks, nil
}

func (t *TaskUseCase) GetTaskByID(ctx context.Context, bucket, taskID string) (*models.TaskEvent, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "get-task-by-id")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucket),
		attribute.String("task-id", taskID),
	)

	taskEvent, err := t.storage.Get(ctx, bucket, taskID)
	if err != nil {
		err = fmt.Errorf("ftask manager error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return taskEvent, nil
}

func (t *TaskUseCase) UpdateTaskStatus(ctx context.Context, task *domain.TaskEvent) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "update-task-status")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", task.ID.String()),
		attribute.String("bucket", task.Bucket),
		attribute.String("message", task.StatusText),
		attribute.Int("status", int(task.Status)),
	)

	taskEventDto := models.FromDomain(task)
	if err := t.storage.Push(ctx, &taskEventDto); err != nil {
		err = fmt.Errorf("task manager error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Warn(err.Error())
	}
}

func (t *TaskUseCase) CheckTaskAlreadyCreated(ctx context.Context, task *domain.TaskEvent) bool {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "check-task-already-created")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", task.ID.String()),
		attribute.String("bucket", task.Bucket),
		attribute.String("message", task.StatusText),
		attribute.Int("status", int(task.Status)),
	)

	storageTask, err := t.storage.Get(ctx, task.Bucket, task.ID.String())
	if err != nil {
		slog.Warn("failed to get task from cache", slog.String("err", err.Error()))
		span.AddEvent("task not found")
		return false
	}

	if storageTask == nil {
		return false
	}

	taskEvent := storageTask.ToDomain()
	switch taskEvent.Status {
	case domain.Received:
		fallthrough
	case domain.Pending:
		fallthrough
	case domain.Processing:
		return true
	case domain.Failed:
		fallthrough
	case domain.Successful:
		return false
	default:
		return false
	}
}

func (t *TaskUseCase) PublishTaskToQueue(ctx context.Context, taskEvent *domain.TaskEvent) error {
	msg := mapping.MessageFromTaskEvent(taskEvent)
	err := t.queue.Publish(ctx, msg)
	if err != nil {
		return fmt.Errorf("task queue error: %w", err)
	}
	return nil
}
