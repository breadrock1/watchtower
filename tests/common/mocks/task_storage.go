package mocks

import (
	"github.com/stretchr/testify/mock"

	"watchtower/internal/shared/kernel"
	"watchtower/internal/support/task/domain"
)

type MockTaskStorage struct {
	mock.Mock
}

func (m *MockTaskStorage) GetTask(_ kernel.Ctx, bucketID kernel.BucketID, taskID kernel.TaskID) (*domain.Task, error) {
	args := m.Called(bucketID, taskID)
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskStorage) GetAllBucketTasks(_ kernel.Ctx, bucketID kernel.BucketID) ([]*domain.Task, error) {
	args := m.Called(bucketID)
	return args.Get(0).([]*domain.Task), args.Error(1)
}

func (m *MockTaskStorage) UpdateTask(_ kernel.Ctx, task *domain.Task) error {
	args := m.Called(task)
	return args.Error(0)
}
