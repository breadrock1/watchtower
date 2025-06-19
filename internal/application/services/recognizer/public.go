package recognizer

import (
	"context"

	"watchtower/internal/application/dto"
)

type IRecognizer interface {
	Recognize(ctx context.Context, inputFile dto.InputFile) (*dto.Recognized, error)
}
