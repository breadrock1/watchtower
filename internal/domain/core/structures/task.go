package domain

import (
	"time"

	"github.com/google/uuid"
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
	Id         uuid.UUID  `json:"id"`
	Bucket     string     `json:"bucket"`
	FilePath   string     `json:"file_path"`
	FileSize   int64      `json:"file_size"`
	StatusText string     `json:"status_text"`
	Status     TaskStatus `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	ModifiedAt time.Time  `json:"modified_at"`
}
