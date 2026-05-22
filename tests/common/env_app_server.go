package common

import (
	"watchtower/cmd"
	"watchtower/cmd/watchtower/httpserver"
	"watchtower/internal/process"
	"watchtower/tests/common/mocks"

	cloudApp "watchtower/internal/core/cloud/application"
	taskApp "watchtower/internal/support/task/application"
)

const (
	TestInstanceKey = "default"
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

	cloudInstances := make(map[string]*cloudApp.StorageUseCase)
	cloudInstances[TestInstanceKey] = storageUseCase
	storagePool := cloudApp.NewStoragePool(cloudInstances)

	taskUseCase := taskApp.NewTaskUseCase(e.TaskStorage, e.TaskQueue, e.Recognizer, e.DocStorage)
	orchestrator := process.NewOrchestrator(servConfig.Orchestrator, storagePool, taskUseCase)
	appServer := httpserver.SetupServer(servConfig.Otlp, orchestrator)
	return appServer, nil
}
