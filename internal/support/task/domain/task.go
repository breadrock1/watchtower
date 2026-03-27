package domain

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/google/uuid"

	"watchtower/internal/shared/kernel"
)

const (
	PublishedStatusText  = "publisher"
	ProcessingStatusText = "processing"
)

// TaskStatus represents the current state of a task in its lifecycle.
// The status follows a typical workflow: Received -> Pending -> Processing -> Successful,
// with Failed as a terminal error state.
type TaskStatus int

const (
	// Failed indicates the task processing encountered an error and could not complete.
	// This is a terminal state.
	Failed TaskStatus = iota - 1 // -1

	// Received indicates the task has been accepted by the queue system
	// but not yet scheduled for processing.
	Received // 0

	// Pending indicates the task is waiting to be processed by a worker.
	Pending // 1

	// Processing indicates the task is currently being executed by a worker.
	Processing // 2

	// Successful indicates the task completed successfully.
	// This is a terminal state.
	Successful // 3
)

// Task represents a unit of work to be processed asynchronously.
// It contains all necessary information for processing and tracks the task's
// lifecycle from creation to completion or failure.
type Task struct {
	// ID uniquely identifies the task across the entire system
	ID kernel.TaskID

	// BucketID identifies which storage bucket contains the input data
	BucketID kernel.BucketID

	// ObjectID identifies the specific object in the bucket to process
	ObjectID kernel.ObjectID

	// ObjectDataSize is the size of the input data in bytes,
	// useful for progress tracking and resource estimation
	ObjectDataSize int

	// StatusText provides additional context about the current status,
	// such as error messages for failed tasks or progress for processing tasks
	StatusText string

	// Status indicates the current state in the task lifecycle
	Status TaskStatus

	// CreatedAt is the timestamp when the task was initially created
	CreatedAt time.Time

	// ModifiedAt is the timestamp of the last status update
	ModifiedAt time.Time

	// RetryCount indicates how many times this task has been retried
	RetryCount int

	// MaxRetries specifies the maximum number of retry attempts
	MaxRetries int

	// ProcessingDuration tracks how long the task took to process (when completed)
	ProcessingDuration time.Duration
}

func CreateNewTask(bucketID kernel.BucketID, objectID kernel.ObjectID) *Task {
	// TODO: Disabled for TechDebt
	// taskID := GenerateUniqID(form.ID, form.FilePath)
	taskID := GenerateTaskID()

	currTime := time.Now()
	task := &Task{
		ID:             taskID,
		BucketID:       bucketID,
		ObjectID:       objectID,
		ObjectDataSize: 0,
		CreatedAt:      currTime,
		ModifiedAt:     currTime,
		Status:         Received,
		StatusText:     PublishedStatusText,
	}

	return task
}

func (t *Task) SetObjectDataSize(size int) {
	t.ObjectDataSize = size
}

func (t *Task) SetStatusAndText(status TaskStatus, msg string) {
	t.Status = status
	t.StatusText = msg
}

func GenerateTaskID() uuid.UUID {
	return uuid.New()
}

func GenerateUniqID(bucket, suffix string) string {
	mask := fmt.Sprintf("%s:%s", bucket, suffix)
	suffix = fmt.Sprintf("%x", md5.Sum([]byte(mask)))
	return suffix
}
