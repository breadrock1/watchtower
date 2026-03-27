package domain

import (
	"bytes"
	"time"
)

// ObjectData represents the actual content of a stored object.
// Using *bytes.Buffer allows efficient reading and writing of object data.
type ObjectData = *bytes.Buffer

// Object represents metadata about a stored object/file in cloud storage.
// It contains all relevant information about the object without its actual data.
type Object struct {
	// Name is the base name of the object (filename)
	// Example: "document.pdf"
	Name string

	// Path is the full path to the object within the bucket
	// Example: "folder/subfolder/document.pdf"
	Path string

	// Checksum is a hash of the object content for integrity verification
	// Usually MD5, SHA256, or provider-specific checksum
	Checksum string

	// ContentType is the MIME type of the object
	// Example: "application/pdf", "image/jpeg"
	ContentType string

	// Expired indicates when the object expires and may be automatically deleted
	// Zero value means the object never expires
	Expired time.Time

	// LastModified is the timestamp of the last modification
	LastModified time.Time

	// Size is the object size in bytes
	Size int64

	// IsDirectory indicates if this "object" actually represents a directory/folder
	// Some storage systems treat folders as objects with special handling
	IsDirectory bool
}
