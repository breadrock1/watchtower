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
	"watchtower/internal/application/utils/telemetry"
	"watchtower/internal/domain/core/structures"
)

const SemaphoreSize = 10

type PipelineUseCase struct {
	processorCh chan models.TaskEvent
	consumerCh  <-chan models.Message

	storageUC    *StorageUseCase
	taskMangerUC *TaskUseCase
	recognizer   recognizer.IRecognizer
}

func NewPipelineUseCase(
	storageUC *StorageUseCase,
	taskMangerUC *TaskUseCase,
	recognizer recognizer.IRecognizer,
) *PipelineUseCase {
	consumerCh := taskMangerUC.GetConsumeChannel()
	processorCh := make(chan models.TaskEvent)
	return &PipelineUseCase{
		processorCh:  processorCh,
		consumerCh:   consumerCh,
		storageUC:    storageUC,
		taskMangerUC: taskMangerUC,
		recognizer:   recognizer,
	}
}

func (p *PipelineUseCase) LaunchListener(ctx context.Context) {
	go func() {
		for {
			select {
			case cMsg := <-p.consumerCh:
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
					p.taskMangerUC.UpdateTaskStatus(ctx, taskEvent)
				}()

			case <-ctx.Done():
				slog.Info("terminating processing")
				return
			}
		}
	}()
}

func (p *PipelineUseCase) CreateTask(ctx context.Context, form models.UploadFileParams) (*domain.TaskEvent, error) {
	task := domain.CreateNewTaskEvent(form.Bucket, form.FilePath, int64(form.FileData.Len()))

	taskID := task.ID.String()
	slog.Info("creating new task",
		slog.String("task-id", taskID),
		slog.String("bucket", form.Bucket),
		slog.String("file-path", form.FilePath),
	)

	ctx, span := telemetry.GlobalTracer.Start(ctx, "create-and-publish-task")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", taskID),
		attribute.String("bucket", task.Bucket),
		attribute.String("file-path", task.FilePath),
		attribute.Int("task-status", int(task.Status)),
		attribute.Int64("time", task.CreatedAt.Unix()),
	)

	if err := p.storageUC.UploadObject(ctx, form); err != nil {
		err = fmt.Errorf("pipeline error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	if err := p.taskMangerUC.PublishTaskToQueue(ctx, task); err != nil {
		err = fmt.Errorf("pipeline error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	p.taskMangerUC.UpdateTaskStatus(ctx, task)
	// TODO: Disabled for TechDebt
	// _ = p.taskMangerUC.CheckTaskAlreadyCreated(ctx, &task)
	// if p.isTaskAlreadyProcessed(ctx, &taskEvent) {
	//	 log.Printf("task has been already processed: %s", taskEvent.ID)
	//	 continue
	// }

	return task, nil
}

func (p *PipelineUseCase) handleTask(ctx context.Context, taskEvent *domain.TaskEvent) {
	slog.Info("processing task event", slog.String("task-id", taskEvent.ID.String()))

	ctx, span := telemetry.GlobalTracer.Start(ctx, "handle-task-from-queue")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", taskEvent.ID.String()),
		attribute.String("bucket", taskEvent.Bucket),
		attribute.String("file-path", taskEvent.FilePath),
	)

	taskEvent.SetStatusAndText(domain.Processing, domain.ProcessingStatusText)
	p.taskMangerUC.UpdateTaskStatus(ctx, taskEvent)

	err := p.processTask(ctx, taskEvent)
	if err != nil {
		err = fmt.Errorf("failed while processing task: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error(err.Error())
		return
	}

	msg := fmt.Sprintf("task %s has been processed successful", taskEvent.ID)
	taskEvent.SetStatusAndText(domain.Successful, msg)
	slog.Info(msg)
}

func (p *PipelineUseCase) processTask(ctx context.Context, taskEvent *domain.TaskEvent) error {
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
		err = fmt.Errorf("failed while processing %s: %w", taskEvent.ID, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	recData, err := p.recognizeObject(ctx, taskEvent, fileData)
	if err != nil {
		taskEvent.SetStatusAndText(domain.Failed, "failed to recognize object data")
		err = fmt.Errorf("failed while processing %s: %w", taskEvent.ID, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	_, err = p.storageUC.StoreDocument(ctx, taskEvent, fileData, recData)
	if err != nil {
		taskEvent.SetStatusAndText(domain.Failed, "failed to store document")
		err = fmt.Errorf("failed while processing %s: %w", taskEvent.ID, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	return nil
}

func (p *PipelineUseCase) loadObject(ctx context.Context, taskEvent *domain.TaskEvent) (bytes.Buffer, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "load-object-from-storage")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", taskEvent.ID.String()),
		attribute.String("bucket", taskEvent.Bucket),
		attribute.String("file-path", taskEvent.FilePath),
	)

	fileData, err := p.storageUC.DownloadObjectByTask(ctx, taskEvent)
	if err != nil {
		taskEvent.SetStatusAndText(domain.Failed, "failed to download file")
		err = fmt.Errorf("failed to download file %s: %w", taskEvent.FilePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return bytes.Buffer{}, err
	}

	return fileData, nil
}

func (p *PipelineUseCase) recognizeObject(
	ctx context.Context,
	taskEvent *domain.TaskEvent,
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
