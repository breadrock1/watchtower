package tokenizer

import (
	"watchtower/internal/application/dto"
)

type ITokenizer interface {
	Load(inputText string) (*dto.Tokens, error)
	LoadByOwnChunked(inputText string) (*dto.Tokens, error)
}
