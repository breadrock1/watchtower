package process

import (
	"context"
	"fmt"
	"log/slog"

	"golang.org/x/sync/semaphore"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"watchtower/internal/core/cloud/domain"
	"watchtower/internal/shared/kernel"
	"watchtower/internal/shared/telemetry"

	cloudApp "watchtower/internal/core/cloud/application"
	taskUC "watchtower/internal/support/task/application"
	taskDomain "watchtower/internal/support/task/domain"
)

type Ctx = context.Context

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

func (o *Orchestrator) LaunchListener(gCtx Ctx) {
	go func() {
		consumeCh := o.taskUC.GetConsumerChannel()
		sem := semaphore.NewWeighted(o.config.SemaphoreSize)
		for {
			select {
			case cMsg := <-consumeCh:
				ctx := cMsg.Ctx
				go func() {
					if err := sem.Acquire(ctx, 1); err != nil {
						slog.Error("internal semaphore error", slog.String("err", err.Error()))
						return
					}
					defer sem.Release(1)

					task := &cMsg.Body
					o.handleTask(ctx, task)
					o.taskUC.UpdateTaskStatus(ctx, task)

					ctx.Done()
				}()

			case <-gCtx.Done():
				slog.Info("terminating processing")
				return
			}
		}
	}()
}

func (o *Orchestrator) UploadFile(
	ctx Ctx,
	bucketID domain.BucketID,
	params *domain.UploadObjectParams,
) (*taskDomain.Task, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "upload-file")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucketID),
		attribute.String("file-path", params.FilePath),
		attribute.Int64("expired", params.Expired.Unix()),
	)

	objID, err := o.storageUC.StoreObject(ctx, bucketID, params)
	if err != nil {
		err = fmt.Errorf("failed to upload file %s: %w", params.FilePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	task, err := o.CreateTask(ctx, bucketID, objID)
	if err != nil {
		err = fmt.Errorf("failed to create taskUC %s: %w", params.FilePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return task, nil
}

func (o *Orchestrator) CreateTask(ctx Ctx, bucketID kernel.BucketID, objID kernel.ObjectID) (*taskDomain.Task, error) {
	task := taskDomain.CreateNewTask(bucketID, objID)

	taskID := task.ID.String()
	slog.Info("creating new task",
		slog.String("task-id", taskID),
		slog.String("bucket", bucketID),
		slog.String("file-path", objID),
	)

	ctx, span := telemetry.GlobalTracer.Start(ctx, "create-and-publish-task")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", taskID),
		attribute.String("bucket", bucketID),
		attribute.String("file-path", objID),
		attribute.Int64("time", task.CreatedAt.Unix()),
		attribute.Int("task-status", int(task.Status)),
	)

	if err := o.taskUC.PublishTaskToQueue(ctx, task); err != nil {
		err = fmt.Errorf("pipeline error: %w", err)
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

func (o *Orchestrator) handleTask(ctx Ctx, task *taskDomain.Task) {
	slog.Info("processing task event", slog.String("task-id", task.ID.String()))

	ctx, span := telemetry.GlobalTracer.Start(ctx, "handle-task-from-queue")
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
		err = fmt.Errorf("pipeline error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Error(err.Error(), slog.String("task-id", task.ID.String()))
		return
	}

	msg := "task has been processed successful"
	task.SetStatusAndText(taskDomain.Successful, msg)
	slog.Info(msg, slog.String("task-id", task.ID.String()))
}

func (o *Orchestrator) processTask(ctx Ctx, task *taskDomain.Task) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "task-processing")
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
