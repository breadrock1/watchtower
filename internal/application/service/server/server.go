package server

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"watchtower/internal/application/usecase"
	"watchtower/internal/application/utils/telemetry"
	"watchtower/internal/domain/core/cloud"
	"watchtower/internal/domain/core/process"
)

type Ctx context.Context

type ServerState struct {
	storageUseCase    *usecase.StorageUseCase
	processingUseCase *usecase.ProcessUseCase
}

func New(objectStorage *usecase.StorageUseCase, taskProcessor *usecase.ProcessUseCase) *ServerState {
	return &ServerState{
		storageUseCase:    objectStorage,
		processingUseCase: taskProcessor,
	}
}

func (s *ServerState) GetObjectStorage() *usecase.StorageUseCase {
	return s.storageUseCase
}

func (s *ServerState) GetTaskProcessor() *usecase.ProcessUseCase {
	return s.processingUseCase
}

func (s *ServerState) UploadFile(
	ctx Ctx,
	bucketID cloud.BucketID,
	params cloud.UploadObjectParams,
) (*process.Task, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "upload-object")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucketID),
		attribute.String("file-path", params.FilePath),
		attribute.Int64("expired", params.Expired.Unix()),
	)

	objID, err := s.storageUseCase.StoreObject(ctx, bucketID, params)
	if err != nil {
		err = fmt.Errorf("failed to upload file %s: %w", params.FilePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	taskParams := &process.CreateTaskParams{BucketID: bucketID, ObjectID: *objID}
	task, err := s.processingUseCase.CreateTask(ctx, taskParams)
	if err != nil {
		err = fmt.Errorf("failed to create task %s: %w", params.FilePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return task, nil
}
