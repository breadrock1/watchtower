package usecase

import (
	"bytes"
	"context"
	"fmt"
	"path"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"watchtower/internal/application/models"
	"watchtower/internal/application/services/storage"
	"watchtower/internal/application/utils/telemetry"
	"watchtower/internal/domain/core/structures"
)

type StorageUseCase struct {
	docStorage storage.IDocumentStorage
	objStorage storage.IObjectStorage
}

func NewStorageUseCase(
	docStorage storage.IDocumentStorage,
	objStorage storage.IObjectStorage,
) *StorageUseCase {
	return &StorageUseCase{docStorage, objStorage}
}

func (s *StorageUseCase) StoreDocument(
	ctx context.Context,
	taskEvent *domain.TaskEvent,
	fileData bytes.Buffer,
	recData *models.Recognized,
) (string, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "store-document-to-index")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", taskEvent.ID.String()),
		attribute.String("bucket", taskEvent.Bucket),
		attribute.String("file-path", taskEvent.FilePath),
	)

	doc := &models.DocumentObject{
		FileName:   path.Base(taskEvent.FilePath),
		FilePath:   taskEvent.FilePath,
		FileSize:   fileData.Len(),
		Content:    recData.Text,
		CreatedAt:  taskEvent.CreatedAt,
		ModifiedAt: taskEvent.ModifiedAt,
	}

	docID, err := s.docStorage.StoreDocument(ctx, taskEvent.Bucket, doc)
	if err != nil {
		err = fmt.Errorf("failed to store document: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return "", err
	}

	return docID, nil
}

func (s *StorageUseCase) GetBuckets(ctx context.Context) ([]models.Bucket, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "get-buckets")
	defer span.End()

	buckets, err := s.objStorage.GetBuckets(ctx)
	if err != nil {
		err = fmt.Errorf("failed to get buckets: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return buckets, err
}

func (s *StorageUseCase) CreateBucket(ctx context.Context, bucket string) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "create-bucket")
	defer span.End()

	span.SetAttributes(attribute.String("bucket", bucket))

	err := s.objStorage.CreateBucket(ctx, bucket)
	if err != nil {
		err = fmt.Errorf("failed to create bucket %s: %w", bucket, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	return nil
}

func (s *StorageUseCase) RemoveBucket(ctx context.Context, bucket string) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "remove-bucket")
	defer span.End()

	span.SetAttributes(attribute.String("bucket", bucket))

	err := s.objStorage.RemoveBucket(ctx, bucket)
	if err != nil {
		err = fmt.Errorf("failed to remove bucket %s: %w", bucket, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}
	return nil
}

func (s *StorageUseCase) IsBucketExists(ctx context.Context, bucket string) (bool, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "is-bucket-exists")
	defer span.End()

	span.SetAttributes(attribute.String("bucket", bucket))

	status, err := s.objStorage.IsBucketExist(ctx, bucket)
	if err != nil {
		err = fmt.Errorf("failed to check if bucket %s exists: %w", bucket, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return false, err
	}

	return status, nil
}

func (s *StorageUseCase) GetFileMetadata(ctx context.Context, bucket, filePath string) (*models.FileAttributes, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "get-file-metadata")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucket),
		attribute.String("file-path", filePath),
	)

	attrs, err := s.objStorage.GetObjectMetadata(ctx, bucket, filePath)
	if err != nil {
		err = fmt.Errorf("failed to get file metadata for %s/%s: %w", bucket, filePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}
	return attrs, nil
}

func (s *StorageUseCase) GetBucketObjects(ctx context.Context, bucket, folder string) ([]models.FileObject, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "get-bucket-objects")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucket),
		attribute.String("folder", folder),
	)

	objects, err := s.objStorage.GetBucketObjects(ctx, bucket, folder)
	if err != nil {
		err := fmt.Errorf("failed to get bucket files: %s/%s", bucket, folder)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return objects, err
}

func (s *StorageUseCase) CopyObject(ctx context.Context, bucket, srcPath, dstPath string) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "copy-object")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucket),
		attribute.String("src-file-path", srcPath),
		attribute.String("dst-file-path", dstPath),
	)

	err := s.objStorage.CopyObject(ctx, bucket, srcPath, dstPath)
	if err != nil {
		err = fmt.Errorf("failed to copy object %s to %s: %w", srcPath, dstPath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}
	return nil
}

func (s *StorageUseCase) DeleteObject(ctx context.Context, bucket, filePath string) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "copy-object")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucket),
		attribute.String("file-path", filePath),
	)

	err := s.objStorage.DeleteObject(ctx, bucket, filePath)
	if err != nil {
		err = fmt.Errorf("failed to remove object %s: %w", filePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}
	return nil
}

func (s *StorageUseCase) MoveObject(ctx context.Context, bucket, srcPath, dstPath string) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "move-file")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucket),
		attribute.String("src-file-path", srcPath),
		attribute.String("dst-file-path", dstPath),
	)

	err := s.objStorage.CopyObject(ctx, bucket, srcPath, dstPath)
	if err != nil {
		err = fmt.Errorf("failed to move object %s to %s: %w", srcPath, dstPath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	err = s.objStorage.DeleteObject(ctx, bucket, dstPath)
	if err != nil {
		err = fmt.Errorf("failed to delete object: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	return nil
}

func (s *StorageUseCase) UploadObject(ctx context.Context, fileForm models.UploadFileParams) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "upload-object")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", fileForm.Bucket),
		attribute.String("file-path", fileForm.FilePath),
		attribute.Int64("expired", fileForm.Expired.Unix()),
		attribute.Int("data-len", fileForm.FileData.Len()),
	)

	err := s.objStorage.UploadObject(ctx, fileForm)
	if err != nil {
		err = fmt.Errorf("failed to upload file %s: %w", fileForm.FilePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	return nil
}

func (s *StorageUseCase) ShareObject(ctx context.Context, params models.ShareObjectParams) (string, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "share-object")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", params.Bucket),
		attribute.String("file-path", params.FilePath),
		attribute.String("expired", params.Expired.String()),
	)

	sharedURL, err := s.objStorage.ShareObjectURL(ctx, params)
	if err != nil {
		err = fmt.Errorf("failed to generate url for %s: %w", params.FilePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return "", err
	}
	return sharedURL, nil
}

func (s *StorageUseCase) DownloadObject(ctx context.Context, bucket, filePath string) (bytes.Buffer, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "download-object")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucket),
		attribute.String("file-path", filePath),
	)

	fileData, err := s.objStorage.DownloadObject(ctx, bucket, filePath)
	if err != nil {
		err = fmt.Errorf("failed to download file %s: %w", filePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return bytes.Buffer{}, err
	}
	return fileData, nil
}

func (s *StorageUseCase) DownloadObjectByTask(ctx context.Context, task *domain.TaskEvent) (bytes.Buffer, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "download-object-by-task")
	defer span.End()

	span.SetAttributes(
		attribute.String("task-id", task.ID.String()),
		attribute.String("bucket", task.Bucket),
		attribute.String("file-path", task.FilePath),
		attribute.Int("task-status", int(task.Status)),
	)

	data, err := s.objStorage.DownloadObject(ctx, task.Bucket, task.FilePath)
	if err != nil {
		err = fmt.Errorf("failed to download file %s: %w", task.FilePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return bytes.Buffer{}, err
	}

	return data, nil
}
