package docparser

import "watchtower/internal/domain/support/recognizer"

type ParsedContent struct {
	Text string `json:"parsed_text"`
}

func (pc *ParsedContent) ToRecognized() recognizer.Recognized {
	return recognizer.Recognized{
		Text: pc.Text,
	}
}
