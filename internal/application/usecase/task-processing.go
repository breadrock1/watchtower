package usecase

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/semaphore"
	"watchtower/internal/application/models"
	"watchtower/internal/application/services/recognizer"
	"watchtower/internal/application/services/task"
	"watchtower/internal/application/utils/telemetry"
	"watchtower/internal/domain/core/structures"
)

const SemaphoreSize = 10

type TaskProcessing struct {
	storageUC   *ObjectStorage
	taskStorage task.ITaskManager
	taskQueue   task.ITaskQueue
	recognizer  recognizer.IRecognizer
}

func NewTaskProcessing(
	storageUC *ObjectStorage,
	taskStorage task.ITaskManager,
	taskQueue task.ITaskQueue,
	recognizer recognizer.IRecognizer,
) *TaskProcessing {
	return &TaskProcessing{
		storageUC:   storageUC,
		taskStorage: taskStorage,
		taskQueue:   taskQueue,
		recognizer:  recognizer,
	}
}

func (t *TaskProcessing) GetAllTasks(ctx context.Context, bucket string) ([]*models.Task, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "load-all-tasks")
	defer span.End()

	span.SetAttributes(attribute.String("bucket", bucket))

	tasks, err := t.taskStorage.GetAll(ctx, bucket)
	if err != nil {
		err = fmt.Errorf("task manager error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return tasks, nil
}

func (t *TaskProcessing) GetTaskByID(ctx context.Context, bucket, taskID string) (*models.Task, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "get-task-by-id")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucket),
		attribute.String("task-id", taskID),
	)

	taskEvent, err := t.taskStorage.Get(ctx, bucket, taskID)
	if err != nil {
		err = fmt.Errorf("ftask manager error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return taskEvent, nil
}

func (t *TaskProcessing) UpdateTaskStatus(ctx context.Context, task *domain.Task) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "update-task-status")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", task.ID.String()),
		attribute.String("bucket", task.Bucket),
		attribute.String("message", task.StatusText),
		attribute.Int("status", int(task.Status)),
	)

	taskEventDto := models.FromDomainTask(task)
	if err := t.taskStorage.Push(ctx, &taskEventDto); err != nil {
		err = fmt.Errorf("task manager error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Warn(err.Error())
	}
}

func (t *TaskProcessing) CheckTaskAlreadyCreated(ctx context.Context, task *domain.Task) bool {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "check-task-already-created")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", task.ID.String()),
		attribute.String("bucket", task.Bucket),
		attribute.String("message", task.StatusText),
		attribute.Int("status", int(task.Status)),
	)

	storageTask, err := t.taskStorage.Get(ctx, task.Bucket, task.ID.String())
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

func (t *TaskProcessing) PublishTaskToQueue(ctx context.Context, taskEvent *domain.Task) error {
	msg := models.MessageFromTask(taskEvent)
	err := t.taskQueue.Publish(ctx, msg)
	if err != nil {
		return fmt.Errorf("task queue error: %w", err)
	}
	return nil
}

func (p *TaskProcessing) LaunchListener(ctx context.Context) {
	go func() {
		consumeCh := p.taskQueue.GetConsumerChannel()
		for {
			select {
			case cMsg := <-consumeCh:
				ctx = cMsg.Ctx

				sem := semaphore.NewWeighted(SemaphoreSize)
				go func() {
					if err := sem.Acquire(ctx, 1); err != nil {
						slog.Error("internal semaphore error", slog.String("err", err.Error()))
						return
					}
					defer sem.Release(1)

					taskEvent := cMsg.Body.ToDomain()
					p.handleTask(ctx, taskEvent)
					p.UpdateTaskStatus(ctx, taskEvent)

					ctx.Done()
				}()

			case <-ctx.Done():
				slog.Info("terminating processing")
				return
			}
		}
	}()
}

func (p *TaskProcessing) CreateTask(ctx context.Context, form models.UploadFileParams) (*domain.Task, error) {
	taskEvent := domain.CreateNewTaskEvent(form.Bucket, form.FilePath, int64(form.FileData.Len()))

	taskID := taskEvent.ID.String()
	slog.Info("creating new task",
		slog.String("task-id", taskID),
		slog.String("bucket", form.Bucket),
		slog.String("file-path", form.FilePath),
	)

	ctx, span := telemetry.GlobalTracer.Start(ctx, "create-and-publish-task")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", taskID),
		attribute.String("bucket", taskEvent.Bucket),
		attribute.String("file-path", taskEvent.FilePath),
		attribute.Int("task-status", int(taskEvent.Status)),
		attribute.Int64("time", taskEvent.CreatedAt.Unix()),
	)

	if err := p.storageUC.UploadObject(ctx, form); err != nil {
		err = fmt.Errorf("pipeline error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	if err := p.PublishTaskToQueue(ctx, taskEvent); err != nil {
		err = fmt.Errorf("pipeline error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	p.UpdateTaskStatus(ctx, taskEvent)
	// TODO: Disabled for TechDebt
	// _ = p.taskMangerUC.CheckTaskAlreadyCreated(ctx, &task)
	// if p.isTaskAlreadyProcessed(ctx, &taskEvent) {
	//	 log.Printf("task has been already processed: %s", taskEvent.ID)
	//	 continue
	// }

	return taskEvent, nil
}

func (p *TaskProcessing) handleTask(ctx context.Context, taskEvent *domain.Task) {
	slog.Info("processing task event", slog.String("task-id", taskEvent.ID.String()))

	ctx, span := telemetry.GlobalTracer.Start(ctx, "handle-task-from-queue")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", taskEvent.ID.String()),
		attribute.String("bucket", taskEvent.Bucket),
		attribute.String("file-path", taskEvent.FilePath),
	)

	taskEvent.SetStatusAndText(domain.Processing, domain.ProcessingStatusText)
	p.UpdateTaskStatus(ctx, taskEvent)

	err := p.processTask(ctx, taskEvent)
	if err != nil {
		err = fmt.Errorf("pipeline error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error(err.Error(), slog.String("task-id", taskEvent.ID.String()))
		return
	}

	msg := "task has been processed successful"
	taskEvent.SetStatusAndText(domain.Successful, msg)
	slog.Info(msg, slog.String("task-id", taskEvent.ID.String()))
}

func (p *TaskProcessing) processTask(ctx context.Context, taskEvent *domain.Task) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "task-processing")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", taskEvent.ID.String()),
		attribute.String("bucket", taskEvent.Bucket),
		attribute.String("file-path", taskEvent.FilePath),
	)

	fileData, err := p.loadObject(ctx, taskEvent)
	if err != nil {
		taskEvent.SetStatusAndText(domain.Failed, "failed to load object from storage")
		err = fmt.Errorf("task processing failed: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	recData, err := p.recognizeObject(ctx, taskEvent, fileData)
	if err != nil {
		taskEvent.SetStatusAndText(domain.Failed, "failed to recognize object data")
		err = fmt.Errorf("task processing failed: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	_, err = p.storageUC.StoreDocument(ctx, taskEvent, fileData, recData)
	if err != nil {
		taskEvent.SetStatusAndText(domain.Failed, "failed to store document")
		err = fmt.Errorf("task processing failed: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	return nil
}

func (p *TaskProcessing) loadObject(ctx context.Context, taskEvent *domain.Task) (bytes.Buffer, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "load-object-from-storage")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", taskEvent.ID.String()),
		attribute.String("bucket", taskEvent.Bucket),
		attribute.String("file-path", taskEvent.FilePath),
	)

	fileData, err := p.storageUC.DownloadObjectByTask(ctx, taskEvent)
	if err != nil {
		err = fmt.Errorf("load object error: %w", err)
		taskEvent.SetStatusAndText(domain.Failed, err.Error())
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return bytes.Buffer{}, err
	}

	return fileData, nil
}

func (p *TaskProcessing) recognizeObject(
	ctx context.Context,
	taskEvent *domain.Task,
	fileData bytes.Buffer,
) (*models.Recognized, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "recognize-object-data")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", taskEvent.ID.String()),
		attribute.String("bucket", taskEvent.Bucket),
		attribute.String("file-path", taskEvent.FilePath),
	)

	inputFile := models.InputFile{
		Data: fileData,
		Name: taskEvent.FilePath,
	}

	recData, err := p.recognizer.Recognize(ctx, inputFile)
	if err != nil {
		taskEvent.SetStatusAndText(domain.Failed, "failed to recognize file")
		err = fmt.Errorf("failed to recognize file %s: %w", taskEvent.ID, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return recData, err
}
