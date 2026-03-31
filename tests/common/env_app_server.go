package common

import (
	"watchtower/cmd"
	"watchtower/cmd/watchtower/httpserver"
	"watchtower/internal/process"
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
	storageUseCase := cloudApp.NewStorageUseCase(e.ObjectStorage)
	taskUseCase := taskApp.NewTaskUseCase(e.TaskStorage, e.TaskQueue, e.Recognizer, e.DocStorage)
	orchestrator := process.NewOrchestrator(servConfig.Orchestrator, storageUseCase, taskUseCase)
	appServer := httpserver.SetupServer(servConfig.Otlp, orchestrator)
	return appServer, nil
}
