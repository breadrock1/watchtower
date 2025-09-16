package s3

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"watchtower/internal/application/dto"
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
	buckets, err := s.mc.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of buckets: %w", err)
	}

	bucketNames := make([]string, len(buckets))
	for index, bucketInfo := range buckets {
		bucketNames[index] = bucketInfo.Name
	}

	return bucketNames, nil
}

func (s *S3Client) CreateBucket(ctx context.Context, bucket string) error {
	opts := minio.MakeBucketOptions{}
	if err := s.mc.MakeBucket(ctx, bucket, opts); err != nil {
		return fmt.Errorf("failed to create bucket %s: %w", bucket, err)
	}
	return nil
}

func (s *S3Client) RemoveBucket(ctx context.Context, bucket string) error {
	if err := s.mc.RemoveBucket(ctx, bucket); err != nil {
		return fmt.Errorf("failed to remove bucket %s: %w", bucket, err)
	}
	return nil
}

func (s *S3Client) IsBucketExist(ctx context.Context, bucket string) (bool, error) {
	result, err := s.mc.BucketExists(ctx, bucket)
	if err != nil {
		return false, fmt.Errorf("failed to check if bucket %s exists: %w", bucket, err)
	}
	return result, nil
}

func (s *S3Client) GetFileMetadata(ctx context.Context, bucket, filePath string) (*dto.FileAttributes, error) {
	opts := minio.StatObjectOptions{}
	stats, err := s.mc.StatObject(ctx, bucket, filePath, opts)
	if err != nil {
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
	opts := minio.ListObjectsOptions{
		UseV1:     true,
		Prefix:    folder,
		Recursive: false,
	}

	if s.mc.IsOffline() {
		return nil, fmt.Errorf("cloud is offline")
	}

	dirObjects := make([]*dto.FileObject, 0)
	for obj := range s.mc.ListObjects(ctx, bucket, opts) {
		if obj.Err != nil {
			slog.Warn("failed to get object from s3: ", slog.String("err", obj.Err.Error()))
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
	opts := minio.RemoveObjectOptions{}
	if err := s.mc.RemoveObject(ctx, bucket, filePath, opts); err != nil {
		return fmt.Errorf("failed to remove object %s: %w", filePath, err)
	}
	return nil
}

func (s *S3Client) CopyFile(ctx context.Context, bucket, srcPath, dstPath string) error {
	srcOpts := minio.CopySrcOptions{Bucket: bucket, Object: srcPath}
	dstOpts := minio.CopyDestOptions{Bucket: bucket, Object: dstPath}
	_, err := s.mc.CopyObject(ctx, dstOpts, srcOpts)
	if err != nil {
		return fmt.Errorf("failed to copy object %s to %s: %w", srcPath, dstPath, err)
	}

	return nil
}

func (s *S3Client) MoveFile(ctx context.Context, bucket, srcPath, dstPath string) error {
	err := s.CopyFile(ctx, bucket, srcPath, dstPath)
	if err != nil {
		return fmt.Errorf("failed to move object %s to %s: %w", srcPath, dstPath, err)
	}

	return s.DeleteFile(ctx, bucket, srcPath)
}

func (s *S3Client) DownloadFile(ctx context.Context, bucket, filePath string) (bytes.Buffer, error) {
	var objBody bytes.Buffer

	opts := minio.GetObjectOptions{}
	obj, err := s.mc.GetObject(ctx, bucket, filePath, opts)
	if err != nil {
		return objBody, fmt.Errorf("failed to get object from s3: %w", err)
	}

	_, err = objBody.ReadFrom(obj)
	if err != nil {
		return objBody, fmt.Errorf("failed to read loaded object from s3: %w", err)
	}

	return objBody, nil
}

func (s *S3Client) UploadFile(
	ctx context.Context,
	bucket, filePath string,
	data *bytes.Buffer,
	expired *time.Time,
) error {
	opts := minio.PutObjectOptions{}
	if expired != nil {
		opts.Expires = *expired
	}

	dataLen := int64(data.Len())
	_, err := s.mc.PutObject(ctx, bucket, filePath, data, dataLen, opts)
	if err != nil {
		return fmt.Errorf("failed to upload file to s3: %w", err)
	}
	return nil
}

func (s *S3Client) GenSharedURL(ctx context.Context, expired time.Duration, bucket, filePath string) (string, error) {
	url, err := s.mc.PresignedGetObject(ctx, bucket, filePath, expired, map[string][]string{})
	if err != nil {
		return "", fmt.Errorf("failed to generate url: %w", err)
	}

	return url.RequestURI(), nil
}
