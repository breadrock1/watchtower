package models

import (
	"bytes"
	"time"
)

type Bucket struct {
	Bucket    string
	Path      string
	CreatedAt time.Time
}

type FileObject struct {
	FileName      string
	DirectoryName string
	IsDirectory   bool
}

type FileAttributes struct {
	SHA256       string
	ContentType  string
	LastModified time.Time
	Size         int64
	Expires      time.Time
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

type DocumentObject struct {
	FileName   string
	FilePath   string
	FileSize   int
	Content    string
	CreatedAt  time.Time
	ModifiedAt time.Time
}
