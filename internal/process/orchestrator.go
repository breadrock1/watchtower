package process

import (
	"fmt"
	"log/slog"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/breadrock1/otlp-go/otlp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"watchtower/internal/core/cloud/domain"
	"watchtower/internal/shared/kernel"
	"watchtower/internal/shared/metrics"

	cloudApp "watchtower/internal/core/cloud/application"
	taskUC "watchtower/internal/support/task/application"
	taskDomain "watchtower/internal/support/task/domain"
)

type Orchestrator struct {
	config      Config
	taskUC      *taskUC.TaskUseCase
	storagePool *cloudApp.StoragePool
}

func NewOrchestrator(config Config, storagePool *cloudApp.StoragePool, taskUC *taskUC.TaskUseCase) *Orchestrator {
	return &Orchestrator{config: config, storagePool: storagePool, taskUC: taskUC}
}

func (o *Orchestrator) GetObjectStorage(orgID kernel.OrganizationID) (*cloudApp.StorageUseCase, error) {
	key := "default"
	if orgID != "" {
		key = orgID
	}

	instance, err := o.storagePool.GetInstance(key)
	if err != nil {
		return nil, fmt.Errorf("could not get storage for orchestrator: %w", err)
	}

	return instance, nil
}

func (o *Orchestrator) GetTaskProcessor() *taskUC.TaskUseCase {
	return o.taskUC
}

func (o *Orchestrator) Health(ctx kernel.Ctx) error {
	instances := append(o.taskUC.GetHealthInstances(), o.storagePool.GetHealthInstances()...)

	group, gCtx := errgroup.WithContext(ctx)
	for _, instance := range instances {
		instance := instance
		group.Go(func() error {
			return instance.Health(gCtx)
		})
	}

	err := group.Wait()
	if err != nil {
		return err
	}

	return nil
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
					acquireStart := time.Now()

					if err := sem.Acquire(ctx, 1); err != nil {
						slog.Error("processing",
							slog.String("msg", "internal semaphore error"),
							slog.String("err", err.Error()),
						)
						return
					}

					metrics.OrchestratorAcquireWaitDurationSeconds.
						WithLabelValues(kernel.AppName).
						Observe(time.Since(acquireStart).Seconds())

					defer sem.Release(1)

					task := &cMsg.Body

					instant := time.Now()
					o.handleTask(ctx, task)

					elapsedTime := time.Since(instant)
					statusInt := strconv.Itoa(int(task.Status))
					metrics.OrchestratorProcessingDurationSeconds.
						WithLabelValues(kernel.AppName, statusInt).
						Observe(elapsedTime.Seconds())

					task.SetProcessingDuration(elapsedTime)
					o.taskUC.UpdateTaskStatus(ctx, task)

					metrics.OrchestratorProcessingCounter.
						WithLabelValues(kernel.AppName, statusInt).
						Inc()

					metrics.ProcessingTasksStatusCounter.
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

	instance, err := o.GetObjectStorage(params.Organization)
	if err != nil {
		err = fmt.Errorf("could not get storage for orchestrator: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	// Track upload file size
	fileSize := params.FileData.Len()
	if fileSize > 0 {
		metrics.UploadFileSizeBytes.
			WithLabelValues(kernel.AppName).
			Observe(float64(fileSize))
	}

	objID, err := instance.StoreObject(ctx, bucketID, params)

	metrics.UploadedFilesCounter.
		WithLabelValues(kernel.AppName, strconv.FormatBool(err != nil)).
		Inc()

	if err != nil {
		err = fmt.Errorf("failed to upload file %s: %w", params.FilePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	task, err := o.CreateTask(ctx, bucketID, objID, params.Organization)

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
	orgID kernel.OrganizationID,
) (*taskDomain.Task, error) {
	task := taskDomain.CreateNewTask(bucketID, objID, orgID)

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

	cTask, err := o.taskUC.GetTask(ctx, task.BucketID, task.ID)
	if cTask != nil {
		span.SetAttributes(attribute.Int("status", int(task.Status)))
		if cTask.Status == taskDomain.Canceled {
			slog.Info("processing",
				slog.String("msg", "task has been canceled"),
				slog.String("task-id", task.ID.String()),
			)
		}
	}

	task.SetStatusAndText(taskDomain.Processing, taskDomain.ProcessingStatusText)
	o.taskUC.UpdateTaskStatus(ctx, task)

	err = o.processTask(ctx, task)
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

	instance, err := o.GetObjectStorage(task.Organization)
	if err != nil {
		err = fmt.Errorf("could not get storage for orchestrator: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	fileData, err := instance.GetObjectData(ctx, task.BucketID, task.ObjectID)
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

// extractFileExtension extracts the file extension without the dot,
// returning "unknown" if no extension is found.
func extractFileExtension(objID kernel.ObjectID) string {
	ext := strings.ToLower(path.Ext(objID))
	if ext == "" {
		return "unknown"
	}
	return strings.TrimPrefix(ext, ".")
}
