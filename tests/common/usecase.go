package common

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"watchtower/cmd/watchtower/config"
	"watchtower/internal/core/cloud/domain"
	"watchtower/internal/core/cloud/infrastructure/s3"
	"watchtower/internal/process"
	"watchtower/internal/shared/telemetry"
	"watchtower/internal/support/task/infrastructure/redis"
	"watchtower/internal/support/task/infrastructure/rmq"
	"watchtower/tests/common/mocks"

	cloudApp "watchtower/internal/core/cloud/application"
	taskApp "watchtower/internal/support/task/application"
)

const TestBucketName = "watchtower-test-bucket"

type TestEnvironment struct {
	Recognizer *mocks.MockRecognizer
	DocStorage *mocks.MockDocStorage

	ObjStorage  *s3.S3Client
	TaskQueue   *rmq.RabbitMQClient
	TaskManager *redis.RedisClient

	TaskUseCase    *taskApp.TaskUseCase
	StorageUseCase *cloudApp.StorageUseCase
	Orchestrator   *process.Orchestrator
}

func InitTestEnvironment(configFilePath string) (*TestEnvironment, error) {
	ctx := context.Background()
	servConfig, err := config.FromFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFilePath, err)
	}

	tracerProvider, _ := telemetry.InitTracer(servConfig.Otlp.Tracer)
	telemetry.GlobalTracer = tracerProvider

	docParser := new(mocks.MockRecognizer)
	docStorage := new(mocks.MockDocStorage)
	objStorage, err := s3.New(&servConfig.Storage.S3)
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

	storageUseCase := cloudApp.NewStorageUseCase(objStorage)
	taskUseCase := taskApp.NewTaskUseCase(taskStorage, taskQueue, docParser, docStorage)
	orchestrator := process.NewOrchestrator(&servConfig.Orchestrator, storageUseCase, taskUseCase)

	testEnvironment := &TestEnvironment{
		Recognizer:     docParser,
		DocStorage:     docStorage,
		ObjStorage:     objStorage,
		TaskQueue:      taskQueue,
		TaskManager:    taskStorage,
		TaskUseCase:    taskUseCase,
		StorageUseCase: storageUseCase,
		Orchestrator:   orchestrator,
	}

	return testEnvironment, nil
}

func CreateUploadFileParams(filePath string) (*domain.UploadObjectParams, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	data := bytes.NewBuffer(fileData)
	expired := time.Now()
	_ = expired.Add(10 * time.Second)

	form := &domain.UploadObjectParams{
		FilePath: filePath,
		FileData: data,
		Expired:  &expired,
	}

	return form, nil
}
