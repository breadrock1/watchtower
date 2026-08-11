package recognizer

import (
	"watchtower/internal/shared/kernel"
)

type IRecognizer interface {
	kernel.IHealth

	Recognize(ctx kernel.Ctx, params *RecognizeParams) (*Recognized, error)
}
