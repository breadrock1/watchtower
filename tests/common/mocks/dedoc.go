package mocks

import (
	"context"

	"watchtower/internal/application/dto"
)

type MockDedocClient struct{}

func (dc *MockDedocClient) Recognize(_ context.Context, inputFile dto.InputFile) (*dto.Recognized, error) {
	recData := &dto.Recognized{
		Text: inputFile.Data.String(),
	}

	return recData, nil
}
