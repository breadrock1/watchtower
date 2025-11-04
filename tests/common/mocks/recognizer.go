package mocks

import (
	"context"

	rec "watchtower/internal/support/task/application/service/recognizer"

	"github.com/stretchr/testify/mock"
)

type MockRecognizer struct {
	mock.Mock
}

func (m *MockRecognizer) Recognize(_ context.Context, params *rec.RecognizeParams) (*rec.Recognized, error) {
	args := m.Called(params)
	return args.Get(0).(*rec.Recognized), args.Error(1)
}
