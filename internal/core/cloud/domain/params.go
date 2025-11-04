package domain

import (
	"bytes"
	"time"
)

type CopyObjectParams struct {
	SourcePath      string
	DestinationPath string
}

type ShareObjectParams struct {
	FilePath string
	Expired  time.Duration
}

type GetObjectsParams struct {
	PrefixPath string
}

type UploadObjectParams struct {
	FilePath string
	FileData *bytes.Buffer
	Expired  *time.Time
}
