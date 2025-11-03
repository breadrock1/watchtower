package cloud

import (
	"bytes"
	"context"
	"net/url"
)

type ICloudStorage interface {
	IBucketManager
	IObjectManager
	IObjectWalker
	IShareManager
}

type IBucketManager interface {
	GetAllBuckets(ctx context.Context) ([]Bucket, error)
	IsBucketExist(ctx context.Context, bucketID BucketID) (bool, error)
	CreateBucket(ctx context.Context, bucketID BucketID) error
	DeleteBucket(ctx context.Context, bucketID BucketID) error
}

type IObjectManager interface {
	GetObjectInfo(ctx context.Context, bucketID BucketID, objID ObjectID) (Object, error)
	GetObjectData(ctx context.Context, bucketID BucketID, objID ObjectID) (bytes.Buffer, error)
	StoreObject(ctx context.Context, bucketID BucketID, params UploadObjectParams) (ObjectID, error)
	CopyObject(ctx context.Context, bucketID BucketID, params CopyObjectParams) error
	DeleteObject(ctx context.Context, bucketID BucketID, objID ObjectID) error
}

type IObjectWalker interface {
	GetBucketObjects(ctx context.Context, bucketID BucketID, params GetObjectsParams) ([]Object, error)
}

type IShareManager interface {
	GenShareURL(ctx context.Context, bucketID BucketID, params ShareObjectParams) (*url.URL, error)
}
