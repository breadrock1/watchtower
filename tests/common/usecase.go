package common

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"
	"watchtower/internal/application/models"
	"watchtower/internal/application/usecase"
	"watchtower/internal/application/utils/telemetry"
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

	PipelineUC    *usecase.PipelineUseCase
	StorageUC     *usecase.StorageUseCase
	TaskManagerUC *usecase.TaskUseCase
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
	err = taskQueue.Consume(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to launch task queue consumer: %w", err)
	}

	taskMangerUC := usecase.NewTaskUseCase(taskStorage, taskQueue)
	storageUC := usecase.NewStorageUseCase(docStorage, objStorage)
	pipelineUC := usecase.NewPipelineUseCase(storageUC, taskMangerUC, recognizer)

	testEnvironment := &TestEnvironment{
		Recognizer:    recognizer,
		DocStorage:    docStorage,
		ObjStorage:    objStorage,
		TaskQueue:     taskQueue,
		TaskManager:   taskStorage,
		PipelineUC:    pipelineUC,
		StorageUC:     storageUC,
		TaskManagerUC: taskMangerUC,
	}

	return testEnvironment, nil
}

func CreateUploadFileParams(filePath string) (*models.UploadFileParams, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	data := bytes.NewBuffer(fileData)
	expired := time.Now()
	_ = expired.Add(10 * time.Second)

	form := &models.UploadFileParams{
		Bucket:   TestBucketName,
		FilePath: filePath,
		FileData: data,
		Expired:  &expired,
	}

	return form, nil
}
