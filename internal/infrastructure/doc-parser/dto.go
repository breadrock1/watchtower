package doc_parser

import "watchtower/internal/application/models"

type Recognized struct {
	Text string `json:"parsed_text"`
}

func (rec *Recognized) ToRecognized() models.Recognized {
	return models.Recognized{
		Text: rec.Text,
	}
}
