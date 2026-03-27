package application

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"watchtower/internal/core/cloud/domain"
	"watchtower/internal/shared/kernel"
	"watchtower/internal/shared/telemetry"
)

type StorageUseCase struct {
	cloudStorage domain.ICloudStorage
}

func NewStorageUseCase(cloudStorage domain.ICloudStorage) *StorageUseCase {
	return &StorageUseCase{cloudStorage: cloudStorage}
}

func (s *StorageUseCase) GetAllBuckets(ctx kernel.Ctx) ([]domain.Bucket, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "get-buckets")
	defer span.End()

	allBuckets, err := s.cloudStorage.GetAllBuckets(ctx)
	if err != nil {
		err = fmt.Errorf("failed to get buckets: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return allBuckets, err
}

func (s *StorageUseCase) CreateBucket(ctx kernel.Ctx, bucketID kernel.BucketID) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "create-bucket")
	defer span.End()

	span.SetAttributes(attribute.String("bucket", bucketID))

	err := s.cloudStorage.CreateBucket(ctx, bucketID)
	if err != nil {
		err = fmt.Errorf("failed to create bucket %s: %w", bucketID, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	return nil
}

func (s *StorageUseCase) DeleteBucket(ctx kernel.Ctx, bucketID kernel.BucketID) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "remove-bucket")
	defer span.End()

	span.SetAttributes(attribute.String("bucket", bucketID))

	err := s.cloudStorage.DeleteBucket(ctx, bucketID)
	if err != nil {
		err = fmt.Errorf("failed to remove bucket %s: %w", bucketID, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}
	return nil
}

func (s *StorageUseCase) IsBucketExists(ctx kernel.Ctx, bucketID kernel.BucketID) (bool, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "is-bucket-exists")
	defer span.End()

	span.SetAttributes(attribute.String("bucket", bucketID))

	status, err := s.cloudStorage.IsBucketExist(ctx, bucketID)
	if err != nil {
		err = fmt.Errorf("failed to check if bucket %s exists: %w", bucketID, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return false, err
	}

	return status, nil
}

func (s *StorageUseCase) GetObjectInfo(
	ctx kernel.Ctx,
	bucketID kernel.BucketID,
	objID kernel.ObjectID,
) (*domain.Object, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "get-file-metadata")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucketID),
		attribute.String("file-path", objID),
	)

	objectInfo, err := s.cloudStorage.GetObjectInfo(ctx, bucketID, objID)
	if err != nil {
		err = fmt.Errorf("failed to get file metadata for %s/%s: %w", bucketID, objID, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return &objectInfo, nil
}

func (s *StorageUseCase) CopyObject(ctx kernel.Ctx, bucketID kernel.BucketID, params *domain.CopyObjectParams) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "copy-object")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucketID),
		attribute.String("src-file-path", params.SourcePath),
		attribute.String("dst-file-path", params.DestinationPath),
	)

	err := s.cloudStorage.CopyObject(ctx, bucketID, params)
	if err != nil {
		err = fmt.Errorf("failed to copy object %s to %s: %w", params.SourcePath, params.DestinationPath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}
	return nil
}

func (s *StorageUseCase) DeleteObject(ctx kernel.Ctx, bucketID kernel.BucketID, objID kernel.ObjectID) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "copy-object")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucketID),
		attribute.String("file-path", objID),
	)

	err := s.cloudStorage.DeleteObject(ctx, bucketID, objID)
	if err != nil {
		err = fmt.Errorf("failed to remove object %s: %w", objID, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}
	return nil
}

func (s *StorageUseCase) DeleteObjects(ctx kernel.Ctx, bucketID kernel.BucketID, prefix string) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "copy-object")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucketID),
		attribute.String("prefix", prefix),
	)

	err := s.cloudStorage.DeleteObjects(ctx, bucketID, prefix)
	if err != nil {
		err = fmt.Errorf("failed to remove objects %s: %w", bucketID, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}
	return nil
}

func (s *StorageUseCase) MoveObject(ctx kernel.Ctx, bucketID kernel.BucketID, params *domain.CopyObjectParams) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "move-file")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucketID),
		attribute.String("src-file-path", params.SourcePath),
		attribute.String("dst-file-path", params.DestinationPath),
	)

	err := s.cloudStorage.CopyObject(ctx, bucketID, params)
	if err != nil {
		err = fmt.Errorf("failed to move object %s to %s: %w", params.SourcePath, params.DestinationPath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	err = s.cloudStorage.DeleteObject(ctx, bucketID, params.DestinationPath)
	if err != nil {
		err = fmt.Errorf("failed to delete object: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	return nil
}

func (s *StorageUseCase) LoadBucketObjects(
	ctx kernel.Ctx,
	bucketID kernel.BucketID,
	params *domain.GetObjectsParams,
) ([]domain.Object, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "get-bucket-objects")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucketID),
		attribute.String("folder", params.PrefixPath),
	)

	objects, err := s.cloudStorage.GetBucketObjects(ctx, bucketID, params)
	if err != nil {
		err := fmt.Errorf("failed to get bucket files: %s/%s", bucketID, params.PrefixPath)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return objects, err
}

func (s *StorageUseCase) StoreObject(
	ctx kernel.Ctx,
	bucketID kernel.BucketID,
	params *domain.UploadObjectParams,
) (kernel.ObjectID, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "upload-object")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucketID),
		attribute.String("file-path", params.FilePath),
		attribute.Int("data-len", params.FileData.Len()),
	)

	objID, err := s.cloudStorage.StoreObject(ctx, bucketID, params)
	if err != nil {
		err = fmt.Errorf("failed to upload file %s: %w", params.FilePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return "", err
	}

	return objID, nil
}

func (s *StorageUseCase) GenShareURL(
	ctx kernel.Ctx,
	bucketID kernel.BucketID,
	params *domain.ShareObjectParams,
) (string, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "share-object")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucketID),
		attribute.String("file-path", params.FilePath),
		attribute.String("expired", params.Expired.String()),
	)

	sharedURL, err := s.cloudStorage.GenShareURL(ctx, bucketID, params)
	if err != nil {
		err = fmt.Errorf("failed to generate url for %s: %w", params.FilePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return "", err
	}
	return sharedURL.RequestURI(), nil
}

func (s *StorageUseCase) GetObjectData(
	ctx kernel.Ctx,
	bucketID kernel.BucketID,
	objID kernel.ObjectID,
) (domain.ObjectData, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "download-object")
	defer span.End()

	span.SetAttributes(
		attribute.String("bucket", bucketID),
		attribute.String("file-path", objID),
	)

	fileData, err := s.cloudStorage.GetObjectData(ctx, bucketID, objID)
	if err != nil {
		err = fmt.Errorf("failed to download file %s: %w", objID, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}
	return fileData, nil
}
