package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"watchtower/internal/application/models"
)

type MockRecognizer struct {
	mock.Mock
}

func (m *MockRecognizer) Recognize(_ context.Context, inputFile models.InputFile) (*models.Recognized, error) {
	args := m.Called(inputFile)
	return args.Get(0).(*models.Recognized), args.Error(1)
}
