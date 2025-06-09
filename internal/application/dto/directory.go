package dto

import "time"

type Directory struct {
	Bucket    string
	Path      string
	CreatedAt time.Time
}

func FromBucketName(dirName string) Directory {
	return Directory{
		Bucket: dirName,
		Path:   "",
	}
}
