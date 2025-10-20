package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"watchtower/internal/domain/core/structures"
)

type MockDocStorage struct {
	mock.Mock
}

func (m *MockDocStorage) StoreDocument(_ context.Context, doc *domain.Document) (string, error) {
	args := m.Called(doc)
	return args.Get(0).(string), args.Error(1)
}
