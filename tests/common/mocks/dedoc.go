package mocks

import (
	"watchtower/internal/application/dto"
)

type MockDedocClient struct{}

func (dc *MockDedocClient) Recognize(inputFile dto.InputFile) (*dto.Recognized, error) {
	recData := &dto.Recognized{
		Text: inputFile.Data.String(),
	}

	return recData, nil
}
