package dto

import (
	"time"
)

type TaskStatus int

const (
	Failed TaskStatus = iota - 1
	Received
	Pending
	Processing
	Successful
)

type TaskEvent struct {
	ID         string     `json:"id"`
	Bucket     string     `json:"bucket"`
	FilePath   string     `json:"file_path"`
	FileSize   int64      `json:"file_size"`
	CreatedAt  time.Time  `json:"created_at"`
	ModifiedAt time.Time  `json:"modified_at"`
	Status     TaskStatus `json:"status"`
	StatusText string     `json:"status_text"`
}
