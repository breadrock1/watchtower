package redis

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"watchtower/internal/support/task/domain"
)

const (
	testBucketID         = "bucket"
	testObjectID         = "file"
	testOrganizationID   = "org"
	testInvalidTaskID    = "bad"
	testDoneStatusText   = "done"
	testRedisCreatedUnix = int64(10)
	testRedisUpdatedUnix = int64(20)
)

func TestRedisDTO(t *testing.T) {
	t.Run("Convert redis value to task", func(t *testing.T) {
		task := domain.CreateNewTask(testBucketID, testObjectID, testOrganizationID)
		value := &RedisValue{
			ID:         task.ID.String(),
			Bucket:     testBucketID,
			FilePath:   testObjectID,
			CreatedAt:  time.Unix(testRedisCreatedUnix, 0).Unix(),
			ModifiedAt: time.Unix(testRedisUpdatedUnix, 0).Unix(),
			Status:     int(domain.Processing),
			StatusText: domain.ProcessingStatusText,
		}

		got, err := value.ConvertToTask()

		assert.NoError(t, err, "failed to convert redis value to task")
		assert.Equal(t, task.ID, got.ID, "unexpected task id")
		assert.Equal(t, testBucketID, got.BucketID, "unexpected bucket id")
		assert.Equal(t, testObjectID, got.ObjectID, "unexpected object id")
		assert.Equal(t, domain.Processing, got.Status, "unexpected task status")
		assert.Equal(t, domain.ProcessingStatusText, got.StatusText, "unexpected status text")
	})

	t.Run("Reject invalid task id", func(t *testing.T) {
		_, err := (&RedisValue{ID: testInvalidTaskID}).ConvertToTask()

		assert.Error(t, err, "expected invalid task id error")
	})

	t.Run("Convert task to redis value", func(t *testing.T) {
		task := domain.CreateNewTask(testBucketID, testObjectID, testOrganizationID)
		task.SetStatusAndText(domain.Successful, testDoneStatusText)

		value := ConvertFromTaskEvent(task)

		assert.Equal(t, task.ID.String(), value.ID, "unexpected redis task id")
		assert.Equal(t, testBucketID, value.Bucket, "unexpected redis bucket")
		assert.Equal(t, testObjectID, value.FilePath, "unexpected redis file path")
		assert.Equal(t, int(domain.Successful), value.Status, "unexpected redis status")
		assert.Equal(t, testDoneStatusText, value.StatusText, "unexpected redis status text")
	})
}
