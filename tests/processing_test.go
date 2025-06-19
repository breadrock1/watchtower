package processing_test

import (
	"bytes"
	"context"
	"os"
	"path"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"watchtower/internal/application/dto"
	"watchtower/internal/application/usecase"
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
	vectorizerServ := &mocks.MockVectorizerClient{}

	redisServ := redis.New(&servConfig.Cacher.Redis)
	rmqServ, initErr := rmq.New(&servConfig.Queue.Rmq)
	assert.NoError(t, initErr, "failed to init rmq client")
	rmqConfig := servConfig.Queue.Rmq
	initErr = rmqServ.CreateExchange(rmqConfig.Exchange)
	assert.NoError(t, initErr, "failed to init rmq exchange")
	initErr = rmqServ.CreateQueue(rmqConfig.Exchange, rmqConfig.QueueName, rmqConfig.RoutingKey)
	assert.NoError(t, initErr, "failed to init rmq queue")
	initErr = rmqServ.Consume(ctx)
	assert.NoError(t, initErr, "failed to start consuming rmq client")

	s3Serv, initErr := s3.New(&servConfig.Cloud.Minio)
	assert.NoError(t, initErr, "failed to init s3 client")
	_ = s3Serv.CreateBucket(ctx, TestBucketName)
	watchDir := dto.FromBucketName(TestBucketName)
	initErr = s3Serv.LaunchWatcher(ctx, []dto.Directory{watchDir})
	assert.NoError(t, initErr, "failed to launch s3 bucket watcher")

	t.Run("Positive pipeline", func(t *testing.T) {
		searcherServ := mocks.NewMockDocSearcherClient()

		cCtx, cancel := context.WithCancel(ctx)
		useCase := usecase.New(
			s3Serv.GetEventsChannel(),
			rmqServ.GetConsumerChannel(),
			rmqServ,
			redisServ,
			s3Serv,
			dedocServ,
			vectorizerServ,
			searcherServ,
		)
		useCase.LaunchProcessing(cCtx)

		fileData, err := os.ReadFile(TestInputFilePath)
		data := bytes.NewBuffer(fileData)
		dataStr := data.String()
		assert.NoError(t, err, "failed to read test input file")
		err = s3Serv.UploadFile(ctx, TestBucketName, path.Base(TestInputFilePath), data)
		assert.NoError(t, err, "failed to upload test input file to s3")

		timeoutCh := time.After(7 * time.Second)
		<-timeoutCh

		task, err := redisServ.Get(ctx, TestBucketName, path.Base(TestInputFilePath))
		assert.NoError(t, err, "failed to get task from redis")
		assert.Equal(t, 3, task.Status)
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
		useCase := usecase.New(
			s3Serv.GetEventsChannel(),
			rmqServ.GetConsumerChannel(),
			rmqServ,
			redisServ,
			s3Serv,
			dedocServ,
			vectorizerServ,
			searcherServ,
		)
		useCase.LaunchProcessing(cCtx)

		taskEvent := dto.TaskEvent{
			Id:         uuid.New(),
			Bucket:     TestBucketName,
			FilePath:   TestInputFilePath,
			FileSize:   0,
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
			Status:     1,
			StatusText: "",
		}

		rmqMsg := dto.FromTaskEvent(taskEvent)
		err := rmqServ.Publish(ctx, rmqMsg)
		assert.NoError(t, err, "failed to publish task event")

		timeoutCh := time.After(7 * time.Second)
		<-timeoutCh

		task, err := redisServ.Get(ctx, TestBucketName, TestInputFilePath)
		assert.NoError(t, err, "failed to get task from redis")
		assert.Equal(t, -1, task.Status)
		assert.Equal(t, TestBucketName, task.Bucket)
		assert.Equal(t, TestInputFilePath, task.FilePath)

		docs := searcherServ.GetDocuments()
		assert.Empty(t, docs, "stored documents is not empty")

		cancel()
	})
}
