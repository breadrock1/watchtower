package processing_test

import (
	"bytes"
	"context"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/jonathanhecl/chunker"
	"github.com/stretchr/testify/assert"
	"watchtower/internal/application/dto"
	"watchtower/internal/application/mapping"
	"watchtower/internal/application/usecase"
	"watchtower/internal/application/utils"
	"watchtower/internal/infrastructure/config"
	"watchtower/internal/infrastructure/redis"
	"watchtower/internal/infrastructure/rmq"
	"watchtower/internal/infrastructure/s3"
	"watchtower/tests/common/mocks"
)

const (
	TestBucketName     = "watchtower-test-bucket"
	TestInputFilePath  = "./resources/input-file.txt"
	TestConfigFilePath = "../configs/testing.toml"
)

func TestProcessing(t *testing.T) {
	var initErr error

	ctx := context.Background()
	servConfig, initErr := config.FromFile(TestConfigFilePath)
	assert.NoError(t, initErr, "failed to load testing config")

	dedocServ := &mocks.MockDedocClient{}

	settings := servConfig.Settings
	textChunker := chunker.NewChunker(
		settings.ChunkSize,
		settings.ChunkOverlap,
		chunker.DefaultSeparators,
		false,
		false,
	)

	redisServ := redis.New(&servConfig.Cacher.Redis)
	rmqServ, initErr := rmq.New(&servConfig.Queue.Rmq)
	assert.NoError(t, initErr, "failed to init rmq client")
	initErr = rmqServ.Consume(ctx)
	assert.NoError(t, initErr, "failed to start consuming rmq client")

	s3Serv, initErr := s3.New(&servConfig.Cloud.S3)
	assert.NoError(t, initErr, "failed to init s3 client")
	_ = s3Serv.CreateBucket(ctx, TestBucketName)

	t.Run("Positive pipeline", func(t *testing.T) {
		searcherServ := mocks.NewMockDocSearcherClient()

		cCtx, cancel := context.WithCancel(ctx)
		useCase := usecase.NewUseCase(
			*textChunker,
			rmqServ,
			redisServ,
			dedocServ,
			searcherServ,
			s3Serv,
		)
		useCase.LaunchWatcherListener(cCtx)

		fileData, err := os.ReadFile(TestInputFilePath)
		data := bytes.NewBuffer(fileData)
		dataStr := strings.Trim(data.String(), "\n")
		assert.NoError(t, err, "failed to read test input file")
		expired := time.Now()
		_ = expired.Add(10 * time.Second)

		fileForm := dto.FileToUpload{
			Bucket:   TestBucketName,
			FilePath: path.Base(TestInputFilePath),
			FileData: data,
			Expired:  &expired,
		}

		task, err := useCase.StoreFileToStorage(ctx, fileForm)
		assert.NoError(t, err, "failed to upload test input file to s3")

		timeoutCh := time.After(7 * time.Second)
		<-timeoutCh

		task, err = redisServ.Get(ctx, TestBucketName, task.ID)
		assert.NoError(t, err, "failed to get task from redis")
		assert.Equal(t, dto.TaskStatus(3), task.Status)
		assert.Equal(t, TestBucketName, task.Bucket)
		assert.Equal(t, path.Base(TestInputFilePath), task.FilePath)

		doc := searcherServ.GetDocuments()[0]
		assert.NotEmpty(t, doc, "stored documents is empty")
		assert.Equal(t, path.Base(TestInputFilePath), doc.FilePath)
		assert.Equal(t, dataStr, doc.Content)

		cancel()
	})

	t.Run("Negative pipeline", func(t *testing.T) {
		searcherServ := mocks.NewMockDocSearcherClient()

		cCtx, cancel := context.WithCancel(ctx)
		useCase := usecase.NewUseCase(
			*textChunker,
			rmqServ,
			redisServ,
			dedocServ,
			searcherServ,
			s3Serv,
		)
		useCase.LaunchWatcherListener(cCtx)

		taskID := utils.GenerateUniqID(TestBucketName, TestInputFilePath)
		taskEvent := dto.TaskEvent{
			ID:         taskID,
			Bucket:     TestBucketName,
			FilePath:   TestInputFilePath,
			FileSize:   0,
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
			Status:     1,
			StatusText: "",
		}

		rmqMsg := mapping.MessageFromTaskEvent(taskEvent)
		err := rmqServ.Publish(ctx, rmqMsg)
		assert.NoError(t, err, "failed to publish task event")

		timeoutCh := time.After(7 * time.Second)
		<-timeoutCh

		task, err := redisServ.Get(ctx, TestBucketName, taskID)
		assert.NoError(t, err, "failed to get task from redis")
		assert.Equal(t, dto.TaskStatus(-1), task.Status)
		assert.Equal(t, TestBucketName, task.Bucket)
		assert.Equal(t, TestInputFilePath, task.FilePath)

		docs := searcherServ.GetDocuments()
		assert.Empty(t, docs, "stored documents is not empty")

		cancel()
	})
}
