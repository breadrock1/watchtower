package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"watchtower/internal/support/task/application/service/recognizer"
)

type MockRecognizer struct {
	mock.Mock
}

func (m *MockRecognizer) Recognize(
	_ context.Context,
	params recognizer.RecognizeParams,
) (recognizer.Recognized, error) {
	args := m.Called(params)
	return args.Get(0).(recognizer.Recognized), args.Error(1)
}
