package dto

import (
	"time"

	"github.com/google/uuid"
)

//go:generate stringer -type=TaskStatus
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
	CreatedAt  time.Time  `json:"created_at"`
	ModifiedAt time.Time  `json:"modified_at"`
	Status     TaskStatus `json:"status"`
	StatusText string     `json:"status_text"`
}

func TaskStatusFromString(enumVal string) TaskStatus {
	switch enumVal {
	case "received":
		return Received
	case "pending":
		return Pending
	case "processing":
		return Processing
	case "successful":
		return Successful
	case "failed":
		return Failed
	default:
		return Pending
	}
}

func (ts *TaskStatus) TaskStatusToInt() int {
	return int(*ts)
}
