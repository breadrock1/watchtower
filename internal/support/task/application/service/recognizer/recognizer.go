package recognizer

import (
	"watchtower/internal/shared/kernel"
)

type IRecognizer interface {
	Recognize(ctx kernel.Ctx, params *RecognizeParams) (*Recognized, error)
}
