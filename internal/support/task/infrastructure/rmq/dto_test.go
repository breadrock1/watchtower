package rmq

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"watchtower/internal/support/task/domain"
)

const (
	testDTOBucketID       = "bucket"
	testDTOObjectID       = "file"
	testDTOOrganizationID = "org"
	testDTOChangedObject  = "changed-file"
	testDTOFailedText     = "failed"
	testDTOObjectDataSize = 128
)

func TestRabbitMQDTO(t *testing.T) {
	t.Run("Convert transport message to domain message", func(t *testing.T) {
		eventID := uuid.New()
		task := *domain.CreateNewTask(testDTOBucketID, testDTOObjectID, testDTOOrganizationID)
		msg := &Message{Ctx: context.Background(), EventId: eventID, Body: task}

		got := msg.ToMessage()

		assert.Equal(t, eventID, got.EventId, "unexpected event id")
		assert.Equal(t, task.ID, got.Body.ID, "unexpected task id")
		assert.NotNil(t, got.Ctx, "expected message context")
	})

	t.Run("Copy full task body", func(t *testing.T) {
		eventID := uuid.New()
		task := *domain.CreateNewTask(testDTOBucketID, testDTOObjectID, testDTOOrganizationID)
		task.SetObjectDataSize(testDTOObjectDataSize)
		task.SetStatusAndText(domain.Processing, domain.ProcessingStatusText)
		task.IncRetryCount()
		msg := &Message{Ctx: context.Background(), EventId: eventID, Body: task}

		got := msg.ToMessage()

		assert.Equal(t, task, got.Body, "expected full task body to be copied")
	})

	t.Run("Keep nil context", func(t *testing.T) {
		eventID := uuid.New()
		task := *domain.CreateNewTask(testDTOBucketID, testDTOObjectID, testDTOOrganizationID)
		msg := &Message{Ctx: nil, EventId: eventID, Body: task}

		got := msg.ToMessage()

		assert.Nil(t, got.Ctx, "expected nil context to be preserved")
		assert.Equal(t, eventID, got.EventId, "unexpected event id")
		assert.Equal(t, task, got.Body, "unexpected task body")
	})

	t.Run("Keep nil event id", func(t *testing.T) {
		task := *domain.CreateNewTask(testDTOBucketID, testDTOObjectID, testDTOOrganizationID)
		msg := &Message{Ctx: context.Background(), EventId: uuid.Nil, Body: task}

		got := msg.ToMessage()

		assert.Equal(t, uuid.Nil, got.EventId, "expected nil event id to be preserved")
		assert.Equal(t, task, got.Body, "unexpected task body")
	})

	t.Run("Keep empty task body", func(t *testing.T) {
		eventID := uuid.New()
		msg := &Message{Ctx: context.Background(), EventId: eventID, Body: domain.Task{}}

		got := msg.ToMessage()

		assert.Equal(t, domain.Task{}, got.Body, "expected empty task body")
		assert.Equal(t, eventID, got.EventId, "unexpected event id")
	})

	t.Run("Keep converted body independent from source changes", func(t *testing.T) {
		eventID := uuid.New()
		task := *domain.CreateNewTask(testDTOBucketID, testDTOObjectID, testDTOOrganizationID)
		msg := &Message{Ctx: context.Background(), EventId: eventID, Body: task}

		got := msg.ToMessage()

		msg.Body.ObjectID = testDTOChangedObject
		msg.Body.SetStatusAndText(domain.Failed, testDTOFailedText)

		assert.Equal(t, testDTOObjectID, got.Body.ObjectID, "expected original object id")
		assert.Equal(t, domain.Received, got.Body.Status, "expected original status")
		assert.Equal(t, domain.PublishedStatusText, got.Body.StatusText, "expected original status text")
	})

	t.Run("Panic on nil transport message", func(t *testing.T) {
		var msg *Message

		assert.Panics(t, func() {
			msg.ToMessage()
		}, "expected panic for nil transport message")
	})

	t.Run("Convert unmarshaled transport message", func(t *testing.T) {
		eventID := uuid.New()
		task := *domain.CreateNewTask(testDTOBucketID, testDTOObjectID, testDTOOrganizationID)
		source := Message{EventId: eventID, Body: task}

		data, err := json.Marshal(source)
		require.NoError(t, err, "expected source message to marshal")

		var decoded Message
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err, "expected source message to unmarshal")

		got := decoded.ToMessage()

		assert.Equal(t, eventID, got.EventId, "unexpected event id")
		assert.Equal(t, task.ID, got.Body.ID, "unexpected task id")
		assert.Equal(t, task.BucketID, got.Body.BucketID, "unexpected bucket id")
		assert.Equal(t, task.ObjectID, got.Body.ObjectID, "unexpected object id")
	})
}
