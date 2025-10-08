package usecase

import (
	"bytes"
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"watchtower/internal/application/dto"
	"watchtower/internal/application/mapping"
	"watchtower/internal/application/services/doc-storage"
	"watchtower/internal/application/services/object-storage"
	"watchtower/internal/application/utils/telemetry"
)

type StorageUseCase struct {
	docStorage doc_storage.IDocumentStorage
	objStorage object_storage.IObjectStorage
}

func NewStorageUseCase(
	docStorage doc_storage.IDocumentStorage,
	objStorage object_storage.IObjectStorage,
) *StorageUseCase {
	return &StorageUseCase{
		docStorage: docStorage,
		objStorage: objStorage,
	}
}

func (suc *StorageUseCase) GetObjectStorage() object_storage.IObjectStorage {
	return suc.objStorage
}

func (suc *StorageUseCase) DownloadObject(ctx context.Context, task *dto.TaskEvent) (bytes.Buffer, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "download-object")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", task.Bucket),
		attribute.String("file-path", task.FilePath),
		attribute.String("task-id", task.ID),
		attribute.Int("task-status", mapping.TaskStatusToInt(task.Status)),
	)

	data, err := suc.objStorage.DownloadFile(ctx, task.Bucket, task.FilePath)
	if err != nil {
		err = fmt.Errorf("failed to download object: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return bytes.Buffer{}, err
	}

	return data, nil
}

func (suc *StorageUseCase) UploadObject(ctx context.Context, fileForm dto.FileToUpload) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "upload-object")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", fileForm.Bucket),
		attribute.String("file-path", fileForm.FilePath),
		attribute.Int64("expired", fileForm.Expired.Unix()),
	)

	err := suc.objStorage.UploadFile(ctx, fileForm.Bucket, fileForm.FilePath, fileForm.FileData, fileForm.Expired)
	if err != nil {
		err = fmt.Errorf("failed to upload object: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	return nil
}

func (suc *StorageUseCase) StoreDocument(ctx context.Context, index string, doc *dto.DocumentObject) (string, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "store-document")
	defer span.End()

	span.SetAttributes(
		attribute.String("index", index),
		attribute.String("file-path", doc.FilePath),
	)

	docID, err := suc.docStorage.StoreDocument(ctx, index, doc)
	if err != nil {
		err = fmt.Errorf("failed to store document: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return "", err
	}

	return docID, nil
}
