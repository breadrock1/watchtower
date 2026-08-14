package mapping

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"watchtower/internal/support/task/domain"
)

const (
	testBucketID       = "bucket"
	testObjectID       = "file"
	testOrganizationID = "org"
	testChangedObject  = "changed-file"
	testFailedText     = "failed"
	testObjectDataSize = 128
)

func TestMessageMapping(t *testing.T) {
	t.Run("Create message from task", func(t *testing.T) {
		task := domain.CreateNewTask(testBucketID, testObjectID, testOrganizationID)

		msg := MessageFromTask(task)

		assert.NotEmpty(t, msg.EventId.String(), "expected generated event id")
		assert.Equal(t, task.ID, msg.Body.ID, "unexpected task id")
		assert.Equal(t, testObjectID, msg.Body.ObjectID, "unexpected object id")
	})

	t.Run("Copy full task body", func(t *testing.T) {
		task := domain.CreateNewTask(testBucketID, testObjectID, testOrganizationID)
		task.SetObjectDataSize(testObjectDataSize)
		task.SetStatusAndText(domain.Processing, domain.ProcessingStatusText)
		task.IncRetryCount()

		msg := MessageFromTask(task)

		assert.Equal(t, *task, msg.Body, "expected message body to contain copied task")
	})

	t.Run("Generate separate event id", func(t *testing.T) {
		task := domain.CreateNewTask(testBucketID, testObjectID, testOrganizationID)

		msg := MessageFromTask(task)

		assert.NotEqual(t, task.ID, msg.EventId, "expected event id to be separate from task id")
	})

	t.Run("Generate unique event id for each message", func(t *testing.T) {
		task := domain.CreateNewTask(testBucketID, testObjectID, testOrganizationID)

		firstMsg := MessageFromTask(task)
		secondMsg := MessageFromTask(task)

		assert.NotEqual(t, firstMsg.EventId, secondMsg.EventId, "expected different event ids")
		assert.Equal(t, firstMsg.Body, secondMsg.Body, "expected same task body")
	})

	t.Run("Keep message body independent from source task changes", func(t *testing.T) {
		task := domain.CreateNewTask(testBucketID, testObjectID, testOrganizationID)

		msg := MessageFromTask(task)

		task.ObjectID = testChangedObject
		task.SetStatusAndText(domain.Failed, testFailedText)

		assert.Equal(t, testObjectID, msg.Body.ObjectID, "expected original object id in message")
		assert.Equal(t, domain.Received, msg.Body.Status, "expected original status in message")
		assert.Equal(t, domain.PublishedStatusText, msg.Body.StatusText, "expected original status text in message")
	})

	t.Run("Accept empty task", func(t *testing.T) {
		task := &domain.Task{}

		msg := MessageFromTask(task)

		assert.NotEmpty(t, msg.EventId.String(), "expected generated event id")
		assert.Equal(t, domain.Task{}, msg.Body, "expected empty task body")
	})

	t.Run("Panic on nil task", func(t *testing.T) {
		assert.Panics(t, func() {
			MessageFromTask(nil)
		}, "expected panic for nil task")
	})

	t.Run("Leave transport metadata empty", func(t *testing.T) {
		task := domain.CreateNewTask(testBucketID, testObjectID, testOrganizationID)

		msg := MessageFromTask(task)

		assert.Nil(t, msg.Ctx, "expected empty context")
		assert.Nil(t, msg.Metadata, "expected empty metadata")
		assert.Zero(t, msg.DeliveryAttempt, "expected zero delivery attempt")
	})
}
