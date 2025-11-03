package httpserver

import (
	"time"

	"watchtower/internal/domain/core/cloud"
	"watchtower/internal/domain/core/process"
)

// TaskSchema example
type TaskSchema struct {
	ID             string    `json:"id"`
	BucketID       string    `json:"bucket_id"`
	ObjectID       string    `json:"object_id"`
	ObjectDataSize int       `json:"object_data_size"`
	StatusText     string    `json:"status_text"`
	Status         int       `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	ModifiedAt     time.Time `json:"modified_at"`
}

func TaskFromDomain(task process.Task) TaskSchema {
	return TaskSchema{
		ID:             task.ID.String(),
		BucketID:       task.BucketID,
		ObjectID:       task.ObjectID,
		ObjectDataSize: task.ObjectDataSize,
		StatusText:     task.StatusText,
		Status:         int(task.Status),
		CreatedAt:      task.CreatedAt,
		ModifiedAt:     task.ModifiedAt,
	}
}

// BucketSchema example
type BucketSchema struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

func BucketFromDomain(bucket cloud.Bucket) BucketSchema {
	return BucketSchema{
		ID:        bucket.ID,
		Path:      bucket.Path,
		CreatedAt: bucket.CreatedAt,
	}
}

// ObjectSchema example
type ObjectSchema struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Checksum     string    `json:"checksum"`
	ContentType  string    `json:"content_type"`
	Expired      time.Time `json:"expired"`
	LastModified time.Time `json:"last_modified"`
	Size         int64     `json:"size"`
	IsDirectory  bool      `json:"is_directory"`
}

func ObjectFromDomain(object cloud.Object) ObjectSchema {
	return ObjectSchema{
		Name:         object.Name,
		Path:         object.Path,
		Checksum:     object.Checksum,
		ContentType:  object.ContentType,
		Expired:      object.Expired,
		LastModified: object.LastModified,
		Size:         object.Size,
		IsDirectory:  object.IsDirectory,
	}
}
