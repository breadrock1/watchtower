package domain

import (
	"crypto/md5"
	"fmt"
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

const PublishedStatusText = "publisher"
const ProcessingStatusText = "processing"

type TaskEvent struct {
	ID         uuid.UUID
	Bucket     string
	FilePath   string
	FileSize   int64
	StatusText string
	Status     TaskStatus
	CreatedAt  time.Time
	ModifiedAt time.Time
}

func GenerateTaskID() uuid.UUID {
	return uuid.New()
}

func GenerateUniqID(bucket, suffix string) string {
	mask := fmt.Sprintf("%s:%s", bucket, suffix)
	suffix = fmt.Sprintf("%x", md5.Sum([]byte(mask)))
	return suffix
}

func CreateNewTaskEvent(bucket, filePath string, fileSize int64) *TaskEvent {
	// TODO: Disabled for TechDebt
	// taskID := GenerateUniqID(form.Bucket, form.FilePath)
	taskID := GenerateTaskID()

	currTime := time.Now()
	task := &TaskEvent{
		ID:         taskID,
		Bucket:     bucket,
		FilePath:   filePath,
		FileSize:   fileSize,
		CreatedAt:  currTime,
		ModifiedAt: currTime,
		Status:     Received,
		StatusText: PublishedStatusText,
	}

	return task
}

func (t *TaskEvent) SetStatusAndText(status TaskStatus, msg string) {
	t.Status = status
	t.StatusText = msg
}
