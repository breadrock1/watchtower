package common

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"watchtower/internal/application/usecase"
	"watchtower/internal/application/utils/telemetry"
	"watchtower/internal/domain/core/cloud"
	"watchtower/internal/infrastructure/config"
	"watchtower/internal/infrastructure/redis"
	"watchtower/internal/infrastructure/rmq"
	"watchtower/internal/infrastructure/s3"
	"watchtower/tests/common/mocks"
)

const TestBucketName = "watchtower-test-bucket"

type TestEnvironment struct {
	Recognizer *mocks.MockRecognizer
	DocStorage *mocks.MockDocStorage

	ObjStorage  *s3.S3Client
	TaskQueue   *rmq.RabbitMQClient
	TaskManager *redis.RedisClient

	TaskProcessing *usecase.ProcessUseCase
	ObjectStorage  *usecase.StorageUseCase
}

func InitTestEnvironment(configFilePath string) (*TestEnvironment, error) {
	ctx := context.Background()
	servConfig, err := config.FromFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFilePath, err)
	}

	tracerProvider, _ := telemetry.InitTracer(servConfig.Server.Tracer)
	telemetry.GlobalTracer = tracerProvider

	recognizer := new(mocks.MockRecognizer)
	docStorage := new(mocks.MockDocStorage)
	objStorage, err := s3.New(&servConfig.Storage.ObjectStorage.S3)
	if err != nil {
		return nil, fmt.Errorf("failed to init object storage: %w", err)
	}
	_ = objStorage.CreateBucket(ctx, TestBucketName)

	taskStorage := redis.New(&servConfig.Task.TaskStorage.Redis)
	taskQueue, err := rmq.New(&servConfig.Task.TaskQueue.Rmq)
	if err != nil {
		return nil, fmt.Errorf("failed to init task queue: %w", err)
	}
	err = taskQueue.StartConsuming(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to launch task queue consumer: %w", err)
	}

	storageUseCase := usecase.NewStorageUseCase(objStorage)
	processingUseCase := usecase.NewProcessingUseCase(objStorage, taskStorage, taskQueue, recognizer, docStorage)

	testEnvironment := &TestEnvironment{
		Recognizer:     recognizer,
		DocStorage:     docStorage,
		ObjStorage:     objStorage,
		TaskQueue:      taskQueue,
		TaskManager:    taskStorage,
		TaskProcessing: processingUseCase,
		ObjectStorage:  storageUseCase,
	}

	return testEnvironment, nil
}

func CreateUploadFileParams(filePath string) (*cloud.UploadObjectParams, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	data := bytes.NewBuffer(fileData)
	expired := time.Now()
	_ = expired.Add(10 * time.Second)

	form := &cloud.UploadObjectParams{
		FilePath: filePath,
		FileData: data,
		Expired:  &expired,
	}

	return form, nil
}
