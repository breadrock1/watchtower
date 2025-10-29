package domain

import "time"

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
