package tokenizer

import (
	"context"

	"watchtower/internal/application/dto"
)

type ITokenizer interface {
	Load(ctx context.Context, inputText string) (*dto.ComputedTokens, error)
	LoadByOwnChunked(ctx context.Context, inputText string) (*dto.ComputedTokens, error)
}
