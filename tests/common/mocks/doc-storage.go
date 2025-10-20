package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"watchtower/internal/application/models"
)

type MockDocStorage struct {
	mock.Mock
}

func (m *MockDocStorage) StoreDocument(_ context.Context, index string, doc *models.DocumentObject) (string, error) {
	var args mock.Arguments
	if doc == nil {
		args = m.Called(index)
	} else {
		args = m.Called(index, doc)
	}

	return args.Get(0).(string), args.Error(1)
}
