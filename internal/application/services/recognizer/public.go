package recognizer

import (
	"context"

	"watchtower/internal/application/models"
)

type IRecognizer interface {
	Recognize(ctx context.Context, inputFile models.InputFile) (*models.Recognized, error)
}
