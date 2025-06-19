package tokenizer

import "watchtower/internal/application/dto"

type ITokenizer interface {
	Load(inputText string) (*dto.ComputedTokens, error)
	LoadByOwnChunked(inputText string) (*dto.ComputedTokens, error)
}
