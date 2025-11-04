package domain

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/google/uuid"
	"watchtower/internal/core/cloud/domain"
)

const (
	PublishedStatusText  = "publisher"
	ProcessingStatusText = "processing"
)

type TaskStatus int

const (
	Failed TaskStatus = iota - 1
	Received
	Pending
	Processing
	Successful
)

type TaskID = uuid.UUID

type Task struct {
	ID             TaskID
	BucketID       domain.BucketID
	ObjectID       domain.ObjectID
	ObjectDataSize int
	StatusText     string
	Status         TaskStatus
	CreatedAt      time.Time
	ModifiedAt     time.Time
}

func CreateNewTask(bucketID domain.BucketID, objectID domain.ObjectID) *Task {
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
