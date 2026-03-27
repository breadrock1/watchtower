package mocks

import (
	"github.com/stretchr/testify/mock"

	"watchtower/internal/shared/kernel"
	"watchtower/internal/support/task/domain"
)

type MockTaskQueue struct {
	mock.Mock

	Ch chan domain.Message
}

func (m *MockTaskQueue) Publish(_ kernel.Ctx, msg domain.Message) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockTaskQueue) GetConsumerChannel() chan domain.Message {
	return m.Ch
}

func (m *MockTaskQueue) StartConsuming(_ kernel.Ctx) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTaskQueue) StopConsuming(_ kernel.Ctx) error {
	args := m.Called()
	return args.Error(0)
}
