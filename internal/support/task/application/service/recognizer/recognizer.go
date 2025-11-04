package recognizer

import (
	"context"
)

type IRecognizer interface {
	Recognize(ctx context.Context, params *RecognizeParams) (*Recognized, error)
}
