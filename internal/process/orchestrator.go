package process

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/breadrock1/otlp-go/otlp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/semaphore"

	"watchtower/internal/core/cloud/domain"
	"watchtower/internal/shared/kernel"
	"watchtower/internal/shared/metrics"

	cloudApp "watchtower/internal/core/cloud/application"
	taskUC "watchtower/internal/support/task/application"
	taskDomain "watchtower/internal/support/task/domain"
)

type Orchestrator struct {
	config    Config
	storageUC *cloudApp.StorageUseCase
	taskUC    *taskUC.TaskUseCase
}

func NewOrchestrator(config Config, storageUC *cloudApp.StorageUseCase, taskUC *taskUC.TaskUseCase) *Orchestrator {
	return &Orchestrator{config: config, storageUC: storageUC, taskUC: taskUC}
}

func (o *Orchestrator) GetObjectStorage() *cloudApp.StorageUseCase {
	return o.storageUC
}

func (o *Orchestrator) GetTaskProcessor() *taskUC.TaskUseCase {
	return o.taskUC
}

func (o *Orchestrator) LaunchListener(ctx kernel.Ctx) {
	slog.Info("starting orchestrator processing")
	go func() {
		consumeCh := o.taskUC.GetConsumerChannel()
		sem := semaphore.NewWeighted(o.config.SemaphoreSize)
		for {
			select {
			case cMsg := <-consumeCh:
				ctx := cMsg.Ctx
				go func() {
					if err := sem.Acquire(ctx, 1); err != nil {
						slog.Error("processing",
							slog.String("msg", "internal semaphore error"),
							slog.String("err", err.Error()),
						)
						return
					}
					defer sem.Release(1)

					task := &cMsg.Body

					instant := time.Now()
					o.handleTask(ctx, task)

					elapsedTime := time.Since(instant)
					statusInt := strconv.Itoa(int(task.Status))
					metrics.OrchestratorProcessingDurationSeconds.
						WithLabelValues(kernel.AppName, statusInt).
						Observe(elapsedTime.Seconds())

					o.taskUC.UpdateTaskStatus(ctx, task)

					metrics.OrchestratorProcessingCounter.
						WithLabelValues(kernel.AppName, statusInt).
						Inc()

					ctx.Done()
				}()

			case <-ctx.Done():
				slog.Info("terminating orchestrator processing")
				return
			}
		}
	}()
}

func (o *Orchestrator) UploadFile(
	ctx kernel.Ctx,
	bucketID kernel.BucketID,
	params *domain.UploadObjectParams,
) (*taskDomain.Task, error) {
	ctx, span := otlp_go.GlobalTracer.Start(ctx, "upload-file")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucketID),
		attribute.String("file-path", params.FilePath),
	)

	objID, err := o.storageUC.StoreObject(ctx, bucketID, params)

	metrics.UploadedFilesCounter.
		WithLabelValues(kernel.AppName, strconv.FormatBool(err != nil)).
		Inc()

	if err != nil {
		err = fmt.Errorf("failed to upload file %s: %w", params.FilePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	task, err := o.CreateTask(ctx, bucketID, objID)

	metrics.CreatedProcessingTasksCounter.
		WithLabelValues(kernel.AppName, strconv.FormatBool(err != nil)).
		Inc()

	if err != nil {
		err = fmt.Errorf("failed to create taskUC %s: %w", params.FilePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return task, nil
}

func (o *Orchestrator) CreateTask(
	ctx kernel.Ctx,
	bucketID kernel.BucketID,
	objID kernel.ObjectID,
) (*taskDomain.Task, error) {
	task := taskDomain.CreateNewTask(bucketID, objID)

	taskID := task.ID.String()
	slog.Info("processing",
		slog.String("msg", "creating new task"),
		slog.String("task-id", taskID),
		slog.String("bucket", bucketID),
		slog.String("file-path", objID),
	)

	ctx, span := otlp_go.GlobalTracer.Start(ctx, "create-and-publish-task")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", taskID),
		attribute.String("bucket", bucketID),
		attribute.String("file-path", objID),
		attribute.Int64("time", task.CreatedAt.Unix()),
		attribute.Int("task-status", int(task.Status)),
	)

	if err := o.taskUC.PublishTaskToQueue(ctx, task); err != nil {
		err = fmt.Errorf("failed to publish task: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	o.taskUC.UpdateTaskStatus(ctx, task)
	// TODO: Disabled for TechDebt
	// _ = p.taskMangerUC.IsTaskAlreadyExists(ctx, &taskDomain)
	// if p.isTaskAlreadyProcessed(ctx, &taskDomain) {
	//	 log.Printf("task has been already processed: %s", taskDomain.ID)
	//	 continue
	// }

	return task, nil
}

func (o *Orchestrator) handleTask(ctx kernel.Ctx, task *taskDomain.Task) {
	slog.Info("processing",
		slog.String("msg", "caught new task"),
		slog.String("task-id", task.ID.String()),
	)

	ctx, span := otlp_go.GlobalTracer.Start(ctx, "handle-task-from-queue")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", task.ID.String()),
		attribute.String("bucket", task.BucketID),
		attribute.String("file-path", task.ObjectID),
	)

	task.SetStatusAndText(taskDomain.Processing, taskDomain.ProcessingStatusText)
	o.taskUC.UpdateTaskStatus(ctx, task)

	err := o.processTask(ctx, task)
	if err != nil {
		err = fmt.Errorf("processing failed: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error("processing",
			slog.String("task-id", task.ID.String()),
			slog.String("err", err.Error()),
		)
		return
	}

	msg := "task has been processed successful"
	task.SetStatusAndText(taskDomain.Successful, msg)
	slog.Info("processing",
		slog.String("msg", msg),
		slog.String("task-id", task.ID.String()),
	)
}

func (o *Orchestrator) processTask(ctx kernel.Ctx, task *taskDomain.Task) error {
	ctx, span := otlp_go.GlobalTracer.Start(ctx, "task-processing")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", task.ID.String()),
		attribute.String("bucket", task.BucketID),
		attribute.String("file-path", task.ObjectID),
	)

	fileData, err := o.storageUC.GetObjectData(ctx, task.BucketID, task.ObjectID)
	if err != nil {
		err = fmt.Errorf("load object error: %w", err)
		task.SetStatusAndText(taskDomain.Failed, err.Error())
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	task.SetObjectDataSize(fileData.Len())
	recData, err := o.taskUC.Recognize(ctx, task, fileData)
	if err != nil {
		task.SetStatusAndText(taskDomain.Failed, "failed to recognize object data")
		err = fmt.Errorf("task processing failed: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	_, err = o.taskUC.StoreDocument(ctx, task, recData)
	if err != nil {
		task.SetStatusAndText(taskDomain.Failed, "failed to store document")
		err = fmt.Errorf("task processing failed: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	return nil
}
