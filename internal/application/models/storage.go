package models

import (
	"bytes"
	"time"

	"watchtower/internal/domain/core/structures"
)

type Bucket struct {
	Name      string
	Path      string
	CreatedAt time.Time
}

func (b *Bucket) ToDomain() domain.Bucket {
	return domain.Bucket{
		Name:      b.Name,
		Path:      b.Path,
		CreatedAt: b.CreatedAt,
	}
}

func FromDomainBucket(bucket *domain.Bucket) *Bucket {
	return &Bucket{
		Name:      bucket.Name,
		Path:      bucket.Path,
		CreatedAt: bucket.CreatedAt,
	}
}

type Object struct {
	Name         string
	Path         string
	Checksum     string
	ContentType  string
	Expired      time.Time
	LastModified time.Time
	Size         int64
	IsDirectory  bool
}

func (o *Object) ToDomain() domain.Object {
	return domain.Object{
		Name:         o.Name,
		Path:         o.Path,
		Checksum:     o.Checksum,
		ContentType:  o.ContentType,
		Expired:      o.Expired,
		LastModified: o.LastModified,
		Size:         o.Size,
		IsDirectory:  o.IsDirectory,
	}
}

func FromDomainObject(obj *domain.Object) *Object {
	return &Object{
		Name:         obj.Name,
		Path:         obj.Path,
		Checksum:     obj.Checksum,
		ContentType:  obj.ContentType,
		Expired:      obj.Expired,
		LastModified: obj.LastModified,
		Size:         obj.Size,
		IsDirectory:  obj.IsDirectory,
	}
}

type UploadFileParams struct {
	Bucket   string
	FilePath string
	FileData *bytes.Buffer
	Expired  *time.Time
}

type ShareObjectParams struct {
	Bucket   string
	FilePath string
	Expired  *time.Duration
}
