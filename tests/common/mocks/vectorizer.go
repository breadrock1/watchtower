package mocks

import (
	"watchtower/internal/application/dto"
)

var (
	InputTextTokens = [][]float64{{0.324254, -1.433421}}
)

type MockVectorizerClient struct{}

func (vc *MockVectorizerClient) Load(inputText string) (*dto.Tokens, error) {
	tokensResult := &dto.Tokens{
		Chunks:      1,
		ChunkedText: []string{inputText},
		Vectors:     InputTextTokens,
	}

	return tokensResult, nil
}

func (vc *MockVectorizerClient) LoadByOwnChunked(inputText string) (*dto.Tokens, error) {
	return vc.Load(inputText)
}
