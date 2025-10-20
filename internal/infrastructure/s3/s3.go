package s3

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"path"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"watchtower/internal/application/models"
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
		return nil, fmt.Errorf("s3 connection error: %w", err)
	}

	client := &S3Client{
		mc: s3Client,
	}

	return client, nil
}

func (s *S3Client) GetBuckets(ctx context.Context) ([]models.Bucket, error) {
	buckets, err := s.mc.ListBuckets(ctx)
	if err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return nil, err
	}

	bucketNames := make([]models.Bucket, len(buckets))
	for index, bucketInfo := range buckets {
		bucketNames[index] = models.Bucket{
			CreatedAt: bucketInfo.CreationDate,
			Bucket:    bucketInfo.Name,
			Path:      "",
		}
	}

	return bucketNames, nil
}

func (s *S3Client) CreateBucket(ctx context.Context, bucket string) error {
	opts := minio.MakeBucketOptions{}
	if err := s.mc.MakeBucket(ctx, bucket, opts); err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return err
	}
	return nil
}

func (s *S3Client) RemoveBucket(ctx context.Context, bucket string) error {
	if err := s.mc.RemoveBucket(ctx, bucket); err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return err
	}
	return nil
}

func (s *S3Client) IsBucketExist(ctx context.Context, bucket string) (bool, error) {
	result, err := s.mc.BucketExists(ctx, bucket)
	if err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return false, err
	}
	return result, nil
}

func (s *S3Client) GetObjectMetadata(ctx context.Context, bucket, filePath string) (*models.FileAttributes, error) {
	opts := minio.StatObjectOptions{}
	filePath = path.Clean(filePath)
	stats, err := s.mc.StatObject(ctx, bucket, filePath, opts)
	if err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return nil, err
	}

	itemAttributes := &models.FileAttributes{
		SHA256:       stats.ChecksumSHA256,
		ContentType:  stats.ContentType,
		LastModified: stats.LastModified,
		Size:         stats.Size,
		Expires:      stats.Expires,
	}

	return itemAttributes, nil
}

func (s *S3Client) GetBucketObjects(ctx context.Context, bucket, folder string) ([]models.FileObject, error) {
	opts := minio.ListObjectsOptions{
		UseV1:     true,
		Prefix:    folder,
		Recursive: false,
	}

	if s.mc.IsOffline() {
		err := fmt.Errorf("s3 connection error")
		return nil, err
	}

	dirObjects := make([]models.FileObject, 0)
	for obj := range s.mc.ListObjects(ctx, bucket, opts) {
		if obj.Err != nil {
			slog.Warn("s3: failed to get object",
				slog.String("bucket", bucket),
				slog.String("err", obj.Err.Error()),
			)
			continue
		}

		dirObjects = append(dirObjects, models.FileObject{
			FileName:      obj.Key,
			DirectoryName: folder,
			IsDirectory:   len(obj.ETag) == 0,
		})
	}

	return dirObjects, nil
}

func (s *S3Client) DeleteObject(ctx context.Context, bucket, filePath string) error {
	opts := minio.RemoveObjectOptions{}
	filePath = path.Clean(filePath)
	if err := s.mc.RemoveObject(ctx, bucket, filePath, opts); err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return err
	}
	return nil
}

func (s *S3Client) CopyObject(ctx context.Context, bucket, srcPath, dstPath string) error {
	srcPath = path.Clean(srcPath)
	dstPath = path.Clean(dstPath)

	srcOpts := minio.CopySrcOptions{Bucket: bucket, Object: srcPath}
	dstOpts := minio.CopyDestOptions{Bucket: bucket, Object: dstPath}
	_, err := s.mc.CopyObject(ctx, dstOpts, srcOpts)
	if err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return err
	}

	return nil
}

func (s *S3Client) DownloadObject(ctx context.Context, bucket, filePath string) (bytes.Buffer, error) {
	var objBody bytes.Buffer

	opts := minio.GetObjectOptions{}
	filePath = path.Clean(filePath)
	obj, err := s.mc.GetObject(ctx, bucket, filePath, opts)
	if err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return objBody, err
	}

	_, err = objBody.ReadFrom(obj)
	if err != nil {
		err = fmt.Errorf("failed while read bytes: %w", err)
		return objBody, err
	}

	return objBody, nil
}

func (s *S3Client) UploadObject(ctx context.Context, params models.UploadFileParams) error {
	opts := minio.PutObjectOptions{}
	if params.Expired != nil {
		opts.Expires = *params.Expired
	}

	dataSize := int64(params.FileData.Len())
	filePath := path.Clean(params.FilePath)
	_, err := s.mc.PutObject(ctx, params.Bucket, filePath, params.FileData, dataSize, opts)
	if err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return err
	}
	return nil
}

func (s *S3Client) ShareObjectURL(ctx context.Context, params models.ShareObjectParams) (string, error) {
	filePath := path.Clean(params.FilePath)
	url, err := s.mc.PresignedGetObject(ctx, params.Bucket, filePath, *params.Expired, map[string][]string{})
	if err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return "", err
	}

	return url.RequestURI(), nil
}
