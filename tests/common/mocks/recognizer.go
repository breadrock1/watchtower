package mocks

import (
	"github.com/stretchr/testify/mock"

	"watchtower/internal/shared/kernel"

	rec "watchtower/internal/support/task/application/service/recognizer"
)

type MockRecognizer struct {
	mock.Mock
}

func (m *MockRecognizer) Recognize(_ kernel.Ctx, params *rec.RecognizeParams) (*rec.Recognized, error) {
	args := m.Called(params)
	return args.Get(0).(*rec.Recognized), args.Error(1)
}
