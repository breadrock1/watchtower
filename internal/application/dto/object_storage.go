package dto

import (
	"bytes"
	"time"
)

type Bucket struct {
	Bucket    string
	Path      string
	CreatedAt time.Time
}

func NewBucketFromName(dirName string) Bucket {
	return Bucket{
		Bucket: dirName,
		Path:   "",
	}
}

type FileObject struct {
	FileName      string `json:"file_name"`
	DirectoryName string `json:"directory_name"`
	IsDirectory   bool   `json:"is_directory"`
}

type FileAttributes struct {
	SHA256       string    `json:"sha256"`
	ContentType  string    `json:"content_type"`
	LastModified time.Time `json:"last_modified"`
	Size         int64     `json:"size"`
	Expires      time.Time `json:"expires"`
}

type FileToUpload struct {
	Bucket   string
	FilePath string
	FileData *bytes.Buffer
	Expired  *time.Time
}
