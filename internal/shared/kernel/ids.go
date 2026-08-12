package kernel

import "github.com/google/uuid"

// OrganizationID is a unique identifier for a external organization cloud.
type OrganizationID = string

// BucketID is a unique identifier for a storage bucket.
// Buckets are top-level containers that hold objects (files).
type BucketID = string

// ObjectID is a unique identifier for an object within a bucket.
// Objects are the actual files stored in the bucket.
type ObjectID = string

// MessageID is a unique identifier for a queue message using UUID v4.
// This is separate from TaskID as the same task might be queued multiple times.
type MessageID = uuid.UUID

// TaskID is a unique identifier for a task using UUID v4.
// This ensures globally unique task identifiers across distributed systems.
type TaskID = uuid.UUID

// CloudInstanceKey is a unique identifier string value for a cloud instance.
type CloudInstanceKey = string
