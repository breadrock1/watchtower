package s3

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/url"
	"path"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"watchtower/internal/core/cloud/domain"
	"watchtower/internal/shared/kernel"
)

type S3Client struct {
	mc *minio.Client
}

func New(config Config) (domain.ICloudStorage, error) {
	creds := credentials.NewStaticV4(config.AccessID, config.SecretKey, config.Token)
	opts := &minio.Options{
		Creds:  creds,
		Secure: config.EnableSSL,
	}

	s3Client, err := minio.New(config.Address, opts)
	if err != nil {
		return nil, fmt.Errorf("s3 connection error: %w", err)
	}

	slog.Info("s3 connection established", slog.String("address", config.Address))

	client := &S3Client{
		mc: s3Client,
	}

	return client, nil
}

func (s *S3Client) GetAllBuckets(ctx kernel.Ctx) ([]domain.Bucket, error) {
	buckets, err := s.mc.ListBuckets(ctx)
	if err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return nil, err
	}

	bucketNames := make([]domain.Bucket, len(buckets))
	for index, bucketInfo := range buckets {
		bucketNames[index] = domain.Bucket{
			CreatedAt: bucketInfo.CreationDate,
			ID:        bucketInfo.Name,
			Path:      "",
		}
	}

	return bucketNames, nil
}

func (s *S3Client) IsBucketExist(ctx kernel.Ctx, bucketID kernel.BucketID) (bool, error) {
	result, err := s.mc.BucketExists(ctx, bucketID)
	if err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return false, err
	}
	return result, nil
}

func (s *S3Client) CreateBucket(ctx kernel.Ctx, bucketID kernel.BucketID) error {
	opts := minio.MakeBucketOptions{}
	if err := s.mc.MakeBucket(ctx, bucketID, opts); err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return err
	}
	return nil
}

func (s *S3Client) DeleteBucket(ctx kernel.Ctx, bucketID kernel.BucketID) error {
	if err := s.mc.RemoveBucket(ctx, bucketID); err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return err
	}
	return nil
}

func (s *S3Client) GetObjectInfo(
	ctx kernel.Ctx,
	bucketID kernel.BucketID,
	objID kernel.ObjectID,
) (domain.Object, error) {
	var objectAttrs domain.Object
	opts := minio.StatObjectOptions{}
	filePath := path.Clean(objID)
	stats, err := s.mc.StatObject(ctx, bucketID, filePath, opts)
	if err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return objectAttrs, err
	}

	objectAttrs = domain.Object{
		Name:         path.Base(filePath),
		Path:         path.Clean(filePath),
		Checksum:     stats.ChecksumSHA256,
		ContentType:  stats.ContentType,
		Expired:      stats.Expires,
		LastModified: stats.LastModified,
		Size:         stats.Size,
		IsDirectory:  len(stats.ETag) == 0,
	}

	return objectAttrs, nil
}

func (s *S3Client) GetObjectData(
	ctx kernel.Ctx,
	bucketID kernel.BucketID,
	objID kernel.ObjectID,
) (domain.ObjectData, error) {
	opts := minio.GetObjectOptions{}
	filePath := path.Clean(objID)
	obj, err := s.mc.GetObject(ctx, bucketID, filePath, opts)
	if err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return nil, err
	}

	objBody := bytes.Buffer{}
	_, err = objBody.ReadFrom(obj)
	if err != nil {
		err = fmt.Errorf("failed while read bytes: %w", err)
		return nil, err
	}

	return &objBody, nil
}

func (s *S3Client) StoreObject(
	ctx kernel.Ctx,
	bucketID kernel.BucketID,
	params *domain.UploadObjectParams,
) (kernel.ObjectID, error) {
	opts := minio.PutObjectOptions{}
	if params.Expired != nil {
		opts.Expires = *params.Expired
	}

	dataSize := int64(params.FileData.Len())
	filePath := path.Clean(params.FilePath)
	_, err := s.mc.PutObject(ctx, bucketID, filePath, params.FileData, dataSize, opts)
	if err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return "", err
	}
	return filePath, nil
}

func (s *S3Client) CopyObject(ctx kernel.Ctx, bucketID kernel.BucketID, params *domain.CopyObjectParams) error {
	srcPath := path.Clean(params.SourcePath)
	dstPath := path.Clean(params.DestinationPath)

	srcOpts := minio.CopySrcOptions{Bucket: bucketID, Object: srcPath}
	dstOpts := minio.CopyDestOptions{Bucket: bucketID, Object: dstPath}
	_, err := s.mc.CopyObject(ctx, dstOpts, srcOpts)
	if err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return err
	}

	return nil
}

func (s *S3Client) DeleteObjects(ctx kernel.Ctx, bucketID kernel.BucketID, prefix string) error {
	listObjOpts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
		UseV1:     true,
	}
	objInfoCh := s.mc.ListObjects(ctx, bucketID, listObjOpts)

	removeObjOpts := minio.RemoveObjectsOptions{
		GovernanceBypass: true,
	}
	errCh := s.mc.RemoveObjects(ctx, bucketID, objInfoCh, removeObjOpts)
	for err := range errCh {
		if err.Err != nil {
			slog.Warn("failed to delete object",
				slog.String("bucket", bucketID),
				slog.String("prefix", prefix),
				slog.String("error", err.ObjectName),
				slog.String("err", err.Err.Error()),
			)
		}
	}

	return nil
}

func (s *S3Client) DeleteObject(ctx kernel.Ctx, bucketID kernel.BucketID, objID kernel.ObjectID) error {
	opts := minio.RemoveObjectOptions{}
	filePath := path.Clean(objID)
	if err := s.mc.RemoveObject(ctx, bucketID, filePath, opts); err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return err
	}
	return nil
}

func (s *S3Client) GetBucketObjects(
	ctx kernel.Ctx,
	bucketID kernel.BucketID,
	params *domain.GetObjectsParams,
) ([]domain.Object, error) {
	opts := minio.ListObjectsOptions{
		Prefix:    params.PrefixPath,
		Recursive: false,
		UseV1:     true,
	}

	if s.mc.IsOffline() {
		err := fmt.Errorf("s3 connection error")
		return nil, err
	}

	dirObjects := make([]domain.Object, 0)
	for obj := range s.mc.ListObjects(ctx, bucketID, opts) {
		if obj.Err != nil {
			slog.Warn("s3: failed to get object",
				slog.String("bucket", bucketID),
				slog.String("err", obj.Err.Error()),
			)
			continue
		}

		dirObjects = append(dirObjects, domain.Object{
			Name:         obj.Key,
			Path:         params.PrefixPath,
			Checksum:     obj.ChecksumSHA256,
			ContentType:  obj.ContentType,
			LastModified: obj.LastModified,
			Expired:      obj.Expiration,
			Size:         obj.Size,
			IsDirectory:  len(obj.ETag) == 0,
		})
	}

	return dirObjects, nil
}

func (s *S3Client) GenShareURL(
	ctx kernel.Ctx,
	bucketID kernel.BucketID,
	params *domain.ShareObjectParams,
) (*url.URL, error) {
	filePath := path.Clean(params.FilePath)
	urlPath, err := s.mc.PresignedGetObject(ctx, bucketID, filePath, params.Expired, map[string][]string{})
	if err != nil {
		err = fmt.Errorf("s3 error: %w", err)
		return nil, err
	}

	return urlPath, nil
}
