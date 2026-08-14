package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	testBucketID            = "bucket"
	testObjectID            = "file"
	testOrganizationID      = "org"
	testEmptyValue          = ""
	testObjectDataSize      = 42
	testRetryCount          = 1
	testFailedStatusMessage = "recognition failed"
	defaultObjectDataSize   = 0
	defaultRetryCount       = 0
	defaultMaxRetries       = 0
	zeroObjectDataSize      = 0
)

const defaultProcessingDuration time.Duration = 0

func TestTask(t *testing.T) {
	t.Run("Create new task", func(t *testing.T) {
		task := CreateNewTask(testBucketID, testObjectID, testOrganizationID)

		assert.NotEmpty(t, task.ID.String(), "expected generated task id")
		assert.Equal(t, testBucketID, task.BucketID, "unexpected bucket id")
		assert.Equal(t, testObjectID, task.ObjectID, "unexpected object id")
		assert.Equal(t, testOrganizationID, task.Organization, "unexpected organization")
		assert.Equal(t, Received, task.Status, "unexpected task status")
		assert.Equal(t, PublishedStatusText, task.StatusText, "unexpected status text")
		assert.False(t, task.CreatedAt.IsZero(), "expected creation time")
		assert.False(t, task.ModifiedAt.IsZero(), "expected modification time")
	})

	t.Run("Update task fields", func(t *testing.T) {
		task := CreateNewTask(testBucketID, testObjectID, testEmptyValue)

		task.SetObjectDataSize(testObjectDataSize)
		task.SetStatusAndText(Processing, ProcessingStatusText)
		task.SetProcessingDuration(time.Second)
		task.IncRetryCount()

		assert.Equal(t, testObjectDataSize, task.ObjectDataSize, "unexpected object data size")
		assert.Equal(t, Processing, task.Status, "unexpected task status")
		assert.Equal(t, ProcessingStatusText, task.StatusText, "unexpected status text")
		assert.Equal(t, time.Second, task.ProcessingDuration, "unexpected processing duration")
		assert.Equal(t, testRetryCount, task.RetryCount, "unexpected retry count")
	})

	t.Run("Generate deterministic id", func(t *testing.T) {
		first := GenerateUniqID(testBucketID, testObjectID)
		second := GenerateUniqID(testBucketID, testObjectID)

		assert.NotEmpty(t, first, "expected generated id")
		assert.Equal(t, first, second, "expected deterministic id")
	})

	t.Run("Create task with default numeric fields", func(t *testing.T) {
		task := CreateNewTask(testBucketID, testObjectID, testOrganizationID)

		assert.Equal(t, defaultObjectDataSize, task.ObjectDataSize, "unexpected default object data size")
		assert.Equal(t, defaultRetryCount, task.RetryCount, "unexpected default retry count")
		assert.Equal(t, defaultMaxRetries, task.MaxRetries, "unexpected default max retries")
		assert.Equal(t, defaultProcessingDuration, task.ProcessingDuration, "unexpected default processing duration")
	})

	t.Run("Create task with equal timestamps", func(t *testing.T) {
		task := CreateNewTask(testBucketID, testObjectID, testOrganizationID)

		assert.Equal(t, task.CreatedAt, task.ModifiedAt, "expected equal initial timestamps")
	})

	t.Run("Generate unique task ids", func(t *testing.T) {
		first := CreateNewTask(testBucketID, testObjectID, testOrganizationID)
		second := CreateNewTask(testBucketID, testObjectID, testOrganizationID)

		assert.NotEqual(t, first.ID, second.ID, "expected different task ids")
	})

	t.Run("Create task with empty identifiers", func(t *testing.T) {
		task := CreateNewTask(testEmptyValue, testEmptyValue, testEmptyValue)

		assert.NotEmpty(t, task.ID.String(), "expected generated task id")
		assert.Equal(t, testEmptyValue, task.BucketID, "unexpected bucket id")
		assert.Equal(t, testEmptyValue, task.ObjectID, "unexpected object id")
		assert.Equal(t, testEmptyValue, task.Organization, "unexpected organization")
		assert.Equal(t, Received, task.Status, "unexpected task status")
	})

	t.Run("Set failed status with error text", func(t *testing.T) {
		task := CreateNewTask(testBucketID, testObjectID, testOrganizationID)

		task.SetStatusAndText(Failed, testFailedStatusMessage)

		assert.Equal(t, Failed, task.Status, "unexpected task status")
		assert.Equal(t, testFailedStatusMessage, task.StatusText, "unexpected status text")
	})

	t.Run("Set zero object data size", func(t *testing.T) {
		task := CreateNewTask(testBucketID, testObjectID, testOrganizationID)

		task.SetObjectDataSize(zeroObjectDataSize)

		assert.Equal(t, zeroObjectDataSize, task.ObjectDataSize, "unexpected object data size")
	})
}
