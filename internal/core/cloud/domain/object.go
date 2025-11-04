package domain

import (
	"bytes"
	"time"
)

type ObjectID = string
type ObjectData = bytes.Buffer

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
