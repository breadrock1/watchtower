package recognizer

import "watchtower/internal/application/dto"

type IRecognizer interface {
	Recognize(inputFile dto.InputFile) (*dto.Recognized, error)
}
