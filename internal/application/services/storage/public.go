package storage

import (
	"bytes"
	"context"

	"watchtower/internal/application/models"
	"watchtower/internal/domain/core/structures"
)

type IObjectStorage interface {
	IBucketManager
	IFileManager
	IFileLoader
	IShareManager
}

type IDocumentStorage interface {
	IDocumentManager
}

type IBucketManager interface {
	GetBuckets(ctx context.Context) ([]models.Bucket, error)
	CreateBucket(ctx context.Context, bucket string) error
	RemoveBucket(ctx context.Context, bucket string) error
	IsBucketExist(ctx context.Context, bucket string) (bool, error)
}

type IFileManager interface {
	DeleteObject(ctx context.Context, bucket, filePath string) error
	CopyObject(ctx context.Context, bucket, srcPath, dstPath string) error
	GetBucketObjects(ctx context.Context, bucket, filePath string) ([]models.FileObject, error)
	GetObjectMetadata(ctx context.Context, bucket, filePath string) (*models.FileAttributes, error)
}

type IShareManager interface {
	ShareObjectURL(ctx context.Context, params models.ShareObjectParams) (string, error)
}

type IFileLoader interface {
	DownloadObject(ctx context.Context, bucket, filePath string) (bytes.Buffer, error)
	UploadObject(ctx context.Context, params models.UploadFileParams) error
}

type IDocumentManager interface {
	StoreDocument(ctx context.Context, document *domain.Document) (string, error)
}
