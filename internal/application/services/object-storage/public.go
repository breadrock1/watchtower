package object_storage

import (
	"bytes"
	"context"
	"time"

	"watchtower/internal/application/dto"
)

type IObjectStorage interface {
	IBucketManager
	IFileManager
	IFileLoader
	IShareManager
}

type IBucketManager interface {
	GetBuckets(ctx context.Context) ([]string, error)
	CreateBucket(ctx context.Context, bucket string) error
	RemoveBucket(ctx context.Context, bucket string) error
	IsBucketExist(ctx context.Context, bucket string) (bool, error)
}

type IFileManager interface {
	DeleteFile(ctx context.Context, bucket, filePath string) error
	CopyFile(ctx context.Context, bucket, srcPath, dstPath string) error
	MoveFile(ctx context.Context, bucket, srcPath, dstPath string) error
	GetBucketFiles(ctx context.Context, bucket, filePath string) ([]*dto.FileObject, error)
	GetFileMetadata(ctx context.Context, bucket, filePath string) (*dto.FileAttributes, error)
}

type IShareManager interface {
	GenSharedURL(ctx context.Context, expired time.Duration, bucket, filePath, redirect string) (string, error)
}

type IFileLoader interface {
	DownloadFile(ctx context.Context, bucket, filePath string) (bytes.Buffer, error)
	UploadFile(ctx context.Context, bucket, filePath string, data *bytes.Buffer, expired *time.Time) error
}
