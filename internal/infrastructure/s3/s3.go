package s3

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"watchtower/internal/application/dto"
	"watchtower/internal/application/utils/telemetry"
)

type S3Client struct {
	mc *minio.Client
}

func New(config *Config) (*S3Client, error) {
	creds := credentials.NewStaticV4(config.AccessID, config.SecretKey, config.Token)
	opts := &minio.Options{
		Creds:  creds,
		Secure: config.EnableSSL,
	}

	s3Client, err := minio.New(config.Address, opts)
	if err != nil {
		return nil, fmt.Errorf("failed while connecting to s3: %w", err)
	}

	client := &S3Client{
		mc: s3Client,
	}

	return client, nil
}

func (s *S3Client) GetBuckets(ctx context.Context) ([]string, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "get-buckets")
	defer span.End()

	span.SetAttributes(
		attribute.String("client", "s3"),
	)

	buckets, err := s.mc.ListBuckets(ctx)
	if err != nil {
		err = fmt.Errorf("failed to get list of buckets: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	bucketNames := make([]string, len(buckets))
	for index, bucketInfo := range buckets {
		bucketNames[index] = bucketInfo.Name
	}

	return bucketNames, nil
}

func (s *S3Client) CreateBucket(ctx context.Context, bucket string) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "create-bucket")
	defer span.End()

	span.SetAttributes(
		attribute.String("client", "s3"),
		attribute.String("bucket", bucket),
	)

	opts := minio.MakeBucketOptions{}
	if err := s.mc.MakeBucket(ctx, bucket, opts); err != nil {
		err = fmt.Errorf("failed to create bucket %s: %w", bucket, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}
	return nil
}

func (s *S3Client) RemoveBucket(ctx context.Context, bucket string) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "remove-bucket")
	defer span.End()

	span.SetAttributes(
		attribute.String("client", "s3"),
		attribute.String("bucket", bucket),
	)

	if err := s.mc.RemoveBucket(ctx, bucket); err != nil {
		err = fmt.Errorf("failed to remove bucket %s: %w", bucket, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}
	return nil
}

func (s *S3Client) IsBucketExist(ctx context.Context, bucket string) (bool, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "is-bucket-exists")
	defer span.End()

	span.SetAttributes(
		attribute.String("client", "s3"),
		attribute.String("bucket", bucket),
	)

	result, err := s.mc.BucketExists(ctx, bucket)
	if err != nil {
		err = fmt.Errorf("failed to check if bucket %s exists: %w", bucket, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return false, err
	}
	return result, nil
}

func (s *S3Client) GetFileMetadata(ctx context.Context, bucket, filePath string) (*dto.FileAttributes, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "get-file-metadata")
	defer span.End()

	span.SetAttributes(
		attribute.String("client", "s3"),
		attribute.String("bucket", bucket),
		attribute.String("file-path", filePath),
	)

	opts := minio.StatObjectOptions{}
	stats, err := s.mc.StatObject(ctx, bucket, filePath, opts)
	if err != nil {
		err = fmt.Errorf("failed to get file metadata for %s/%s: %w", bucket, filePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	itemAttributes := &dto.FileAttributes{
		SHA256:       stats.ChecksumSHA256,
		ContentType:  stats.ContentType,
		LastModified: stats.LastModified,
		Size:         stats.Size,
		Expires:      stats.Expires,
	}

	return itemAttributes, nil
}

func (s *S3Client) GetBucketFiles(ctx context.Context, bucket, folder string) ([]*dto.FileObject, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "get-bucket-files")
	defer span.End()

	span.SetAttributes(
		attribute.String("client", "s3"),
		attribute.String("bucket", bucket),
		attribute.String("folder", folder),
	)

	opts := minio.ListObjectsOptions{
		UseV1:     true,
		Prefix:    folder,
		Recursive: false,
	}

	if s.mc.IsOffline() {
		err := fmt.Errorf("failed to get bucket files: %s/%s", bucket, folder)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	dirObjects := make([]*dto.FileObject, 0)
	for obj := range s.mc.ListObjects(ctx, bucket, opts) {
		if obj.Err != nil {
			slog.Warn("failed to get object", slog.String("err", obj.Err.Error()))
			continue
		}

		dirObjects = append(dirObjects, &dto.FileObject{
			FileName:      obj.Key,
			DirectoryName: folder,
			IsDirectory:   len(obj.ETag) == 0,
		})
	}

	return dirObjects, nil
}

func (s *S3Client) DeleteFile(ctx context.Context, bucket, filePath string) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "delete-file")
	defer span.End()

	span.SetAttributes(
		attribute.String("client", "s3"),
		attribute.String("bucket", bucket),
		attribute.String("file-path", filePath),
	)

	opts := minio.RemoveObjectOptions{}
	if err := s.mc.RemoveObject(ctx, bucket, filePath, opts); err != nil {
		err = fmt.Errorf("failed to remove object %s: %w", filePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}
	return nil
}

func (s *S3Client) CopyFile(ctx context.Context, bucket, srcPath, dstPath string) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "copy-file")
	defer span.End()

	span.SetAttributes(
		attribute.String("client", "s3"),
		attribute.String("bucket", bucket),
		attribute.String("src-file-path", srcPath),
		attribute.String("dst-file-path", dstPath),
	)

	srcOpts := minio.CopySrcOptions{Bucket: bucket, Object: srcPath}
	dstOpts := minio.CopyDestOptions{Bucket: bucket, Object: dstPath}
	_, err := s.mc.CopyObject(ctx, dstOpts, srcOpts)
	if err != nil {
		err = fmt.Errorf("failed to copy object %s to %s: %w", srcPath, dstPath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	return nil
}

func (s *S3Client) MoveFile(ctx context.Context, bucket, srcPath, dstPath string) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "move-file")
	defer span.End()

	span.SetAttributes(
		attribute.String("client", "s3"),
		attribute.String("bucket", bucket),
		attribute.String("src-file-path", srcPath),
		attribute.String("dst-file-path", dstPath),
	)

	err := s.CopyFile(ctx, bucket, srcPath, dstPath)
	if err != nil {
		err = fmt.Errorf("failed to move object %s to %s: %w", srcPath, dstPath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	return s.DeleteFile(ctx, bucket, srcPath)
}

func (s *S3Client) DownloadFile(ctx context.Context, bucket, filePath string) (bytes.Buffer, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "s3-download-file")
	defer span.End()

	span.SetAttributes(
		attribute.String("client", "s3"),
		attribute.String("bucket", bucket),
		attribute.String("file-path", filePath),
	)

	var objBody bytes.Buffer

	opts := minio.GetObjectOptions{}
	obj, err := s.mc.GetObject(ctx, bucket, filePath, opts)
	if err != nil {
		err = fmt.Errorf("failed to download file %s: %w", filePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return objBody, err
	}

	_, err = objBody.ReadFrom(obj)
	if err != nil {
		err = fmt.Errorf("failed to download file %s: %w", filePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return objBody, err
	}

	return objBody, nil
}

func (s *S3Client) UploadFile(
	ctx context.Context,
	bucket, filePath string,
	data *bytes.Buffer,
	expired *time.Time,
) error {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "s3-upload-file")
	defer span.End()

	span.SetAttributes(
		attribute.String("client", "s3"),
		attribute.String("bucket", bucket),
		attribute.String("file-path", filePath),
		attribute.Int("data-len", data.Len()),
	)

	opts := minio.PutObjectOptions{}
	if expired != nil {
		opts.Expires = *expired
	}

	dataLen := int64(data.Len())
	_, err := s.mc.PutObject(ctx, bucket, filePath, data, dataLen, opts)
	if err != nil {
		err = fmt.Errorf("failed to upload file %s: %w", filePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}
	return nil
}

func (s *S3Client) GenSharedURL(ctx context.Context, expired time.Duration, bucket, filePath string) (string, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "share-url")
	defer span.End()

	span.SetAttributes(
		attribute.String("client", "s3"),
		attribute.String("bucket", bucket),
		attribute.String("file-path", filePath),
		attribute.Float64("expired-secs", expired.Seconds()),
	)

	url, err := s.mc.PresignedGetObject(ctx, bucket, filePath, expired, map[string][]string{})
	if err != nil {
		err = fmt.Errorf("failed to generate url for %s: %w", filePath, err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return "", err
	}

	return url.RequestURI(), nil
}
