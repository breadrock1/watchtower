package usecase

import (
	"bytes"
	"fmt"
	"log/slog"
	"path"

	"golang.org/x/sync/semaphore"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"watchtower/internal/application/mapping"
	"watchtower/internal/application/utils/telemetry"
	"watchtower/internal/domain/core/cloud"
	"watchtower/internal/domain/core/process"
	"watchtower/internal/domain/support/docstorage"
	"watchtower/internal/domain/support/recognizer"
)

const SemaphoreSize = 10

type ProcessUseCase struct {
	objectStorage cloud.IObjectManager
	taskStorage   process.ITaskManager
	taskQueue     process.ITaskQueue
	recognizer    recognizer.IRecognizer
	docStorage    docstorage.IDocumentStorage
}

func NewProcessingUseCase(
	objectStorage cloud.IObjectManager,
	taskStorage process.ITaskManager,
	taskQueue process.ITaskQueue,
	recognizer recognizer.IRecognizer,
	docStorage docstorage.IDocumentStorage,
) *ProcessUseCase {
	return &ProcessUseCase{
		objectStorage: objectStorage,
		taskStorage:   taskStorage,
		taskQueue:     taskQueue,
		recognizer:    recognizer,
		docStorage:    docStorage,
	}
}

func (p *ProcessUseCase) GetBucketTasks(ctx Ctx, bucketID cloud.BucketID) ([]*process.Task, error) {
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

func (p *ProcessUseCase) GetTask(ctx Ctx, bucketID cloud.BucketID, taskID process.TaskID) (*process.Task, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "get-task-by-id")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucketID),
		attribute.String("task-id", taskID.String()),
	)

	task, err := p.taskStorage.GetTask(ctx, bucketID, taskID)
	if err != nil {
		err = fmt.Errorf("ftask manager error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return task, nil
}

func (p *ProcessUseCase) UpdateTaskStatus(ctx Ctx, task *process.Task) {
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

func (p *ProcessUseCase) IsTaskAlreadyExists(ctx Ctx, task *process.Task) bool {
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
	case process.Received:
		fallthrough
	case process.Pending:
		fallthrough
	case process.Processing:
		return true
	case process.Failed:
		fallthrough
	case process.Successful:
		return false
	default:
		return false
	}
}

func (p *ProcessUseCase) PublishTaskToQueue(ctx Ctx, task *process.Task) error {
	msg := mapping.MessageFromTask(task)
	err := p.taskQueue.Publish(ctx, msg)
	return err
}

func (p *ProcessUseCase) LaunchListener(ctx Ctx) {
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

					task := &cMsg.Body
					p.handleTask(ctx, task)
					p.UpdateTaskStatus(ctx, task)

					ctx.Done()
				}()

			case <-ctx.Done():
				slog.Info("terminating processing")
				return
			}
		}
	}()
}

func (p *ProcessUseCase) CreateTask(ctx Ctx, params *process.CreateTaskParams) (*process.Task, error) {
	task := process.CreateNewTask(params.BucketID, params.ObjectID)

	taskID := task.ID.String()
	slog.Info("creating new task",
		slog.String("task-id", taskID),
		slog.String("bucket", params.BucketID),
		slog.String("file-path", params.ObjectID),
	)

	ctx, span := telemetry.GlobalTracer.Start(ctx, "create-and-publish-task")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", taskID),
		attribute.String("bucket", params.BucketID),
		attribute.String("file-path", params.ObjectID),
		attribute.Int64("time", task.CreatedAt.Unix()),
		attribute.Int("task-status", int(task.Status)),
	)

	if err := p.PublishTaskToQueue(ctx, task); err != nil {
		err = fmt.Errorf("pipeline error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	p.UpdateTaskStatus(ctx, task)
	// TODO: Disabled for TechDebt
	// _ = p.taskMangerUC.IsTaskAlreadyExists(ctx, &task)
	// if p.isTaskAlreadyProcessed(ctx, &task) {
	//	 log.Printf("task has been already processed: %s", task.ID)
	//	 continue
	// }

	return task, nil
}

func (p *ProcessUseCase) handleTask(ctx Ctx, task *process.Task) {
	slog.Info("processing task event", slog.String("task-id", task.ID.String()))

	ctx, span := telemetry.GlobalTracer.Start(ctx, "handle-task-from-queue")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", task.ID.String()),
		attribute.String("bucket", task.BucketID),
		attribute.String("file-path", task.ObjectID),
	)

	task.SetStatusAndText(process.Processing, process.ProcessingStatusText)
	p.UpdateTaskStatus(ctx, task)

	err := p.processTask(ctx, task)
	if err != nil {
		err = fmt.Errorf("pipeline error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error(err.Error(), slog.String("task-id", task.ID.String()))
		return
	}

	msg := "task has been processed successful"
	task.SetStatusAndText(process.Successful, msg)
	slog.Info(msg, slog.String("task-id", task.ID.String()))
}

func (p *ProcessUseCase) processTask(ctx Ctx, task *process.Task) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "task-processing")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", task.ID.String()),
		attribute.String("bucket", task.BucketID),
		attribute.String("file-path", task.ObjectID),
	)

	fileData, err := p.loadObject(ctx, task)
	if err != nil {
		task.SetStatusAndText(process.Failed, "failed to load object from storage")
		err = fmt.Errorf("task processing failed: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	task.SetObjectDataSize(fileData.Len())
	recData, err := p.recognizeObject(ctx, task, &fileData)
	if err != nil {
		task.SetStatusAndText(process.Failed, "failed to recognize object data")
		err = fmt.Errorf("task processing failed: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	_, err = p.storeDocument(ctx, task, recData)
	if err != nil {
		task.SetStatusAndText(process.Failed, "failed to store document")
		err = fmt.Errorf("task processing failed: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	return nil
}

func (p *ProcessUseCase) loadObject(ctx Ctx, task *process.Task) (bytes.Buffer, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "load-object-from-storage")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", task.ID.String()),
		attribute.String("bucket", task.BucketID),
		attribute.String("file-path", task.ObjectID),
	)

	fileData, err := p.objectStorage.GetObjectData(ctx, task.BucketID, task.ObjectID)
	if err != nil {
		err = fmt.Errorf("load object error: %w", err)
		task.SetStatusAndText(process.Failed, err.Error())
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return bytes.Buffer{}, err
	}

	return fileData, nil
}

func (p *ProcessUseCase) recognizeObject(
	ctx Ctx,
	task *process.Task,
	fileData *bytes.Buffer,
) (recognizer.Recognized, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "recognize-object-data")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", task.ID.String()),
		attribute.String("bucket", task.BucketID),
		attribute.String("file-path", task.ObjectID),
	)

	inputFile := recognizer.RecognizeParams{
		FileName: task.ObjectID,
		FileData: fileData,
	}

	recData, err := p.recognizer.Recognize(ctx, inputFile)
	if err != nil {
		task.SetStatusAndText(process.Failed, "failed to recognize file")
		err = fmt.Errorf("failed to recognize file %s: %w", task.ID, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return recData, err
	}

	return recData, err
}

func (p *ProcessUseCase) storeDocument(
	ctx Ctx,
	task *process.Task,
	recData recognizer.Recognized,
) (docstorage.DocumentID, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "store-document-to-index")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", task.ID.String()),
		attribute.String("bucket", task.BucketID),
		attribute.String("file-path", task.ObjectID),
	)

	doc := docstorage.Document{
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
