package mocks

import (
	"watchtower/internal/shared/kernel"

	"watchtower/internal/support/task/application/service/docstorage"

	"github.com/stretchr/testify/mock"
)

type MockDocStorage struct {
	mock.Mock
}

func (m *MockDocStorage) StoreDocument(_ kernel.Ctx, doc *docstorage.Document) (docstorage.DocumentID, error) {
	args := m.Called(doc)
	return args.Get(0).(string), args.Error(1)
}
