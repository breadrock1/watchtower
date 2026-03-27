package domain

import (
	"bytes"
	"time"
)

// CopyObjectParams defines parameters for copying an object from one location to another
// within the same bucket or across different paths.
type CopyObjectParams struct {
	// SourcePath is the full path of the source object
	// Example: "documents/original/file.pdf"
	SourcePath string

	// DestinationPath is the full path where the object should be copied
	// Example: "backups/file.pdf" or "documents/copy/file.pdf"
	DestinationPath string
}

// ShareObjectParams defines parameters for generating a shareable URL for an object.
// This enables temporary access to private objects.
type ShareObjectParams struct {
	// FilePath is the path to the object to share
	// Example: "shared/report.pdf"
	FilePath string

	// Expired specifies how long the shareable URL remains valid
	// Example: 24 * time.Hour for a 24-hour link
	Expired time.Duration
}

// GetObjectsParams defines filtering and pagination parameters for listing objects.
type GetObjectsParams struct {
	// PrefixPath filters objects to those with paths starting with this prefix
	// This effectively lists objects in a "directory"
	// Example: "documents/2024/" to list all objects in the 2024 folder
	PrefixPath string

	// MaxKeys limits the number of objects returned (pagination)
	// Zero means use provider default
	MaxKeys int32

	// ContinuationToken for pagination through large result sets
	ContinuationToken string
}

// UploadObjectParams defines parameters for uploading a new object to storage.
type UploadObjectParams struct {
	// FilePath is the destination path for the uploaded object
	// Example: "uploads/images/profile.jpg"
	FilePath string

	// FileData contains the actual content to upload
	FileData *bytes.Buffer

	// ContentType specifies the MIME type (optional, auto-detected if not provided)
	ContentType string

	// Expired sets an expiration time for the object (optional)
	// If nil, the object never expires
	Expired *time.Time

	// Metadata allows attaching custom key-value pairs to the object
	Metadata map[string]string
}
