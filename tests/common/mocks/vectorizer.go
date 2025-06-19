package mocks

import (
	"context"

	"watchtower/internal/application/dto"
)

var (
	InputTextTokens = [][]float64{{0.324254, -1.433421}}
)

type MockVectorizerClient struct{}

func (vc *MockVectorizerClient) Load(_ context.Context, inputText string) (*dto.ComputedTokens, error) {
	tokensResult := &dto.ComputedTokens{
		ChunksCount: 1,
		ChunkedText: []string{inputText},
		Vectors:     InputTextTokens,
	}

	return tokensResult, nil
}

func (vc *MockVectorizerClient) LoadByOwnChunked(ctx context.Context, inputText string) (*dto.ComputedTokens, error) {
	return vc.Load(ctx, inputText)
}
