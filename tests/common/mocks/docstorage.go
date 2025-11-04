package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"watchtower/internal/support/task/application/service/docstorage"
)

type MockDocStorage struct {
	mock.Mock
}

func (m *MockDocStorage) StoreDocument(_ context.Context, doc docstorage.Document) (docstorage.DocumentID, error) {
	args := m.Called(doc)
	return args.Get(0).(string), args.Error(1)
}
