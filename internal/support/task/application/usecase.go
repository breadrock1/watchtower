package application

import (
	"bytes"
	"fmt"
	"log/slog"
	"path"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"watchtower/internal/shared/kernel"
	"watchtower/internal/shared/telemetry"
	"watchtower/internal/support/task/application/mapping"
	"watchtower/internal/support/task/application/service/docstorage"
	"watchtower/internal/support/task/application/service/recognizer"
	"watchtower/internal/support/task/domain"
)

type TaskUseCase struct {
	taskStorage domain.ITaskManager
	taskQueue   domain.ITaskQueue
	recognizer  recognizer.IRecognizer
	docStorage  docstorage.IDocumentStorage
}

func NewTaskUseCase(
	taskStorage domain.ITaskManager,
	taskQueue domain.ITaskQueue,
	recognizer recognizer.IRecognizer,
	docStorage docstorage.IDocumentStorage,
) *TaskUseCase {
	return &TaskUseCase{
		taskStorage: taskStorage,
		taskQueue:   taskQueue,
		recognizer:  recognizer,
		docStorage:  docStorage,
	}
}

func (p *TaskUseCase) GetBucketTasks(ctx kernel.Ctx, bucketID kernel.BucketID) ([]*domain.Task, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "get-all-bucket-tasks")
	defer span.End()

	span.SetAttributes(attribute.String("bucket", bucketID))

	allTasks, err := p.taskStorage.GetAllBucketTasks(ctx, bucketID)
	if err != nil {
		err = fmt.Errorf("task manager error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return allTasks, nil
}

func (p *TaskUseCase) GetTask(ctx kernel.Ctx, bucketID kernel.BucketID, taskID kernel.TaskID) (*domain.Task, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "get-task-by-id")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucketID),
		attribute.String("task-id", taskID.String()),
	)

	task, err := p.taskStorage.GetTask(ctx, bucketID, taskID)
	if err != nil {
		err = fmt.Errorf("task manager error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return task, nil
}

func (p *TaskUseCase) UpdateTaskStatus(ctx kernel.Ctx, task *domain.Task) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "update-task-status")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", task.ID.String()),
		attribute.String("bucket", task.BucketID),
		attribute.String("message", task.StatusText),
		attribute.Int("status", int(task.Status)),
	)

	if err := p.taskStorage.UpdateTask(ctx, task); err != nil {
		err = fmt.Errorf("task manager error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Warn(err.Error())
	}
}

func (p *TaskUseCase) IsTaskAlreadyExists(ctx kernel.Ctx, task *domain.Task) bool {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "check-task-already-created")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", task.ID.String()),
		attribute.String("bucket", task.BucketID),
		attribute.String("message", task.StatusText),
		attribute.Int("status", int(task.Status)),
	)

	task, err := p.taskStorage.GetTask(ctx, task.BucketID, task.ID)
	if err != nil {
		slog.Warn("failed to get task from cache", slog.String("err", err.Error()))
		span.AddEvent("task not found")
		return false
	}

	if task == nil {
		return false
	}

	switch task.Status {
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

func (p *TaskUseCase) PublishTaskToQueue(ctx kernel.Ctx, task *domain.Task) error {
	msg := mapping.MessageFromTask(task)
	err := p.taskQueue.Publish(ctx, msg)
	return err
}

func (p *TaskUseCase) GetConsumerChannel() chan domain.Message {
	return p.taskQueue.GetConsumerChannel()
}

func (p *TaskUseCase) Recognize(
	ctx kernel.Ctx,
	task *domain.Task,
	fileData *bytes.Buffer,
) (*recognizer.Recognized, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "recognize-object-data")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", task.ID.String()),
		attribute.String("bucket", task.BucketID),
		attribute.String("file-path", task.ObjectID),
	)

	inputFile := &recognizer.RecognizeParams{
		FileName: task.ObjectID,
		FileData: fileData,
	}

	// TODO: impled retry pattern
	recData, err := p.recognizer.Recognize(ctx, inputFile)
	if err != nil {
		task.SetStatusAndText(domain.Failed, "failed to recognize file")
		err = fmt.Errorf("failed to recognize file %s: %w", task.ID, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return recData, err
	}

	return recData, err
}

func (p *TaskUseCase) StoreDocument(
	ctx kernel.Ctx,
	task *domain.Task,
	recData *recognizer.Recognized,
) (docstorage.DocumentID, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "store-document-to-index")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", task.ID.String()),
		attribute.String("bucket", task.BucketID),
		attribute.String("file-path", task.ObjectID),
	)

	doc := &docstorage.Document{
		Index:      task.BucketID,
		Name:       path.Base(task.ObjectID),
		Path:       task.ObjectID,
		Size:       task.ObjectDataSize,
		Content:    recData.Text,
		CreatedAt:  task.CreatedAt,
		ModifiedAt: task.ModifiedAt,
	}

	docID, err := p.docStorage.StoreDocument(ctx, doc)
	if err != nil {
		err = fmt.Errorf("failed to store document: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return "", err
	}

	return docID, nil
}
