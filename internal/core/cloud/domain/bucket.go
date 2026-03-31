package domain

import (
	"time"
	"watchtower/internal/shared/kernel"
)

// Bucket represents a storage bucket/container in the cloud storage system.
// Buckets are used to organize objects and control access at the container level.
type Bucket struct {
	// ID is the unique identifier for the bucket
	ID kernel.BucketID

	// Path is the full path or URI to the bucket
	// Example: "s3://my-bucket" or "gs://my-bucket"
	Path string

	// CreatedAt indicates when the bucket was created
	CreatedAt time.Time
}
