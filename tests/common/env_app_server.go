package common

import (
	"fmt"
	"watchtower/cmd"
	"watchtower/cmd/watchtower/httpserver"
	"watchtower/internal/process"
	"watchtower/internal/shared/telemetry"
	"watchtower/tests/common/mocks"

	cloudApp "watchtower/internal/core/cloud/application"
	taskApp "watchtower/internal/support/task/application"
)

type TestAppServerEnvironment struct {
	ObjectStorage *mocks.MockObjectStorage
	TaskStorage   *mocks.MockTaskStorage
	TaskQueue     *mocks.MockTaskQueue
	DocStorage    *mocks.MockDocStorage
	Recognizer    *mocks.MockRecognizer
}

func InitTestAppEnvironment() *TestAppServerEnvironment {
	objectStorage := new(mocks.MockObjectStorage)
	taskStorage := new(mocks.MockTaskStorage)
	taskQueue := new(mocks.MockTaskQueue)
	recognizer := new(mocks.MockRecognizer)
	docStorage := new(mocks.MockDocStorage)
	return &TestAppServerEnvironment{
		ObjectStorage: objectStorage,
		TaskStorage:   taskStorage,
		TaskQueue:     taskQueue,
		DocStorage:    docStorage,
		Recognizer:    recognizer,
	}
}

func (e *TestAppServerEnvironment) BuildAppServer(servConfig *cmd.Config) (*httpserver.Server, error) {
	tracerProvider, err := telemetry.InitTracer(servConfig.Otlp.Tracer)
	telemetry.GlobalTracer = tracerProvider
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracer: %w", err)
	}

	storageUseCase := cloudApp.NewStorageUseCase(e.ObjectStorage)
	taskUseCase := taskApp.NewTaskUseCase(e.TaskStorage, e.TaskQueue, e.Recognizer, e.DocStorage)
	orchestrator := process.NewOrchestrator(servConfig.Orchestrator, storageUseCase, taskUseCase)
	appServer := httpserver.SetupServer(servConfig.Otlp, orchestrator, tracerProvider)
	return appServer, nil
}
