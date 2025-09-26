package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"watchtower/internal/application/dto"
	"watchtower/internal/application/mapping"
	"watchtower/internal/application/services/task-manager"
	"watchtower/internal/application/utils/telemetry"
)

type TaskMangerUseCase struct {
	taskManager task_manager.ITaskManager
}

func NewTaskManagerUseCase(taskManager task_manager.ITaskManager) *TaskMangerUseCase {
	return &TaskMangerUseCase{
		taskManager: taskManager,
	}
}

func (tuc *TaskMangerUseCase) GetTaskManager() task_manager.ITaskManager {
	return tuc.taskManager
}

func (tuc *TaskMangerUseCase) UpdateTaskStatus(ctx context.Context, task *dto.TaskEvent) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "update-task-status")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", task.Bucket),
		attribute.String("task-id", task.ID),
		attribute.String("message", task.StatusText),
		attribute.Int("status", mapping.TaskStatusToInt(task.Status)),
	)

	if err := tuc.taskManager.Push(ctx, task); err != nil {
		err = fmt.Errorf("failed to update task: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Warn(err.Error())
	}
}

func (tuc *TaskMangerUseCase) CheckTaskAlreadyCreated(ctx context.Context, task *dto.TaskEvent) bool {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "check-task-already-created")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", task.Bucket),
		attribute.String("task-id", task.ID),
		attribute.String("message", task.StatusText),
		attribute.Int("status", mapping.TaskStatusToInt(task.Status)),
	)

	storageTask, err := tuc.taskManager.Get(ctx, task.Bucket, task.ID)
	if err != nil {
		slog.Warn("failed to get task from cache", slog.String("err", err.Error()))
		span.AddEvent("task not found")
		return false
	}

	if storageTask == nil {
		return false
	}

	switch storageTask.Status {
	case dto.Received:
		fallthrough
	case dto.Pending:
		fallthrough
	case dto.Processing:
		return true
	case dto.Failed:
		fallthrough
	case dto.Successful:
		return false
	default:
		return false
	}
}
