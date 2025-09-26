package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"watchtower/internal/application/dto"
	"watchtower/internal/application/mapping"
	"watchtower/internal/application/services/recognizer"
	"watchtower/internal/application/services/task-queue"
	"watchtower/internal/application/utils"
	"watchtower/internal/application/utils/telemetry"
)

type PipelineUseCase struct {
	processorCh chan dto.TaskEvent
	consumerCh  <-chan dto.Message

	storageUC    *StorageUseCase
	taskMangerUC *TaskMangerUseCase
	taskQueue    task_queue.ITaskQueue
	recognizer   recognizer.IRecognizer
}

func NewPipelineUseCase(
	storageUC *StorageUseCase,
	taskMangerUC *TaskMangerUseCase,
	taskQueue task_queue.ITaskQueue,
	recognizer recognizer.IRecognizer,
) *PipelineUseCase {
	consumerCh := taskQueue.GetConsumerChannel()
	processorCh := make(chan dto.TaskEvent)
	return &PipelineUseCase{
		processorCh:  processorCh,
		consumerCh:   consumerCh,
		storageUC:    storageUC,
		taskMangerUC: taskMangerUC,
		taskQueue:    taskQueue,
		recognizer:   recognizer,
	}
}

func (puc *PipelineUseCase) LaunchWatcherListener(ctx context.Context) {
	go func() {
		for {
			select {
			case cMsg := <-puc.consumerCh:
				ctx = cMsg.Ctx
				task := puc.handleConsumedTask(ctx, cMsg)
				puc.taskMangerUC.UpdateTaskStatus(ctx, &task)
			case <-ctx.Done():
				slog.Info("terminating processing")
				return
			}
		}
	}()
}

func (puc *PipelineUseCase) CreateAndPublishTask(ctx context.Context, fileForm dto.FileToUpload) (*dto.TaskEvent, error) {
	// TODO: Disabled for TechDebt
	// taskID := utils.GenerateUniqID(fileForm.Bucket, fileForm.FilePath)
	taskID := utils.GenerateTaskID()

	slog.Info("created new task",
		slog.String("task-taskID", taskID),
		slog.String("bucket", fileForm.Bucket),
		slog.String("file-path", fileForm.FilePath),
	)

	currTime := time.Now()
	task := dto.TaskEvent{
		ID:         taskID,
		Bucket:     fileForm.Bucket,
		FilePath:   fileForm.FilePath,
		FileSize:   int64(fileForm.FileData.Len()),
		CreatedAt:  currTime,
		ModifiedAt: currTime,
		Status:     dto.Received,
	}

	ctx, span := telemetry.GlobalTracer.Start(ctx, "create-and-publish-task")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", task.Bucket),
		attribute.String("file-path", task.FilePath),
		attribute.String("task-id", taskID),
		attribute.Int("task-status", mapping.TaskStatusToInt(task.Status)),
		attribute.Int64("time", currTime.Unix()),
	)

	if err := puc.storageUC.UploadObject(ctx, fileForm); err != nil {
		err = fmt.Errorf("failed to upload file: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	if err := puc.publishTaskToQueue(ctx, task); err != nil {
		err = fmt.Errorf("failed to publish task to queue: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	puc.taskMangerUC.UpdateTaskStatus(ctx, &task)
	// TODO: Disabled for TechDebt
	// _ = puc.taskMangerUC.CheckTaskAlreadyCreated(ctx, &task)
	// if puc.isTaskAlreadyProcessed(ctx, &taskEvent) {
	//	 log.Printf("task has been already processed: %s", taskEvent.ID)
	//	 continue
	// }

	return &task, nil
}

func (puc *PipelineUseCase) publishTaskToQueue(ctx context.Context, taskEvent dto.TaskEvent) error {
	msg := mapping.MessageFromTaskEvent(taskEvent)
	err := puc.taskQueue.Publish(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to publish task event to queue: %w", err)
	}
	return nil
}

func (puc *PipelineUseCase) handleConsumedTask(ctx context.Context, recvMessage dto.Message) dto.TaskEvent {
	taskEvent := recvMessage.Body
	slog.Info("processing task event", slog.String("task-id", taskEvent.ID))

	ctx, span := telemetry.GlobalTracer.Start(ctx, "handle-task-from-queue")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", taskEvent.ID),
		attribute.String("bucket", taskEvent.Bucket),
		attribute.String("file-path", taskEvent.FilePath),
	)

	taskEvent.Status = dto.Processing
	taskEvent.StatusText = dto.EmptyMessage
	puc.taskMangerUC.UpdateTaskStatus(ctx, &taskEvent)

	err := puc.processTaskEvent(ctx, recvMessage.Body)
	if err != nil {
		err = fmt.Errorf("failed while processing file: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		taskEvent.Status = dto.Failed
		taskEvent.StatusText = err.Error()

		return taskEvent
	}

	msg := fmt.Sprintf("task %s has been processed successful", taskEvent.ID)
	slog.Info(msg)
	taskEvent.StatusText = msg
	taskEvent.Status = dto.Successful

	return taskEvent
}

func (puc *PipelineUseCase) processTaskEvent(ctx context.Context, taskEvent dto.TaskEvent) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "task-processing")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", taskEvent.ID),
		attribute.String("bucket", taskEvent.Bucket),
		attribute.String("file-path", taskEvent.FilePath),
	)

	fileData, err := puc.storageUC.DownloadObject(ctx, taskEvent)
	if err != nil {
		err = fmt.Errorf("failed to download file: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	inputFile := dto.InputFile{
		Data: fileData,
		Name: path.Base(taskEvent.FilePath),
	}

	recData, err := puc.recognizer.Recognize(ctx, inputFile)
	if err != nil {
		err = fmt.Errorf("failed to recognize file: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	doc := &dto.DocumentObject{
		FileName:   path.Base(taskEvent.FilePath),
		FilePath:   taskEvent.FilePath,
		FileSize:   fileData.Len(),
		Content:    recData.Text,
		CreatedAt:  taskEvent.CreatedAt,
		ModifiedAt: taskEvent.ModifiedAt,
	}

	docID, err := puc.storageUC.StoreDocument(ctx, taskEvent.Bucket, doc)
	if err != nil {
		err = fmt.Errorf("failed to store document %s: %w", doc.FileName, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	slog.Debug("document has been stored",
		slog.String("task-id", taskEvent.ID),
		slog.String("bucket", taskEvent.Bucket),
		slog.String("document-id", docID),
	)

	return nil
}
