package dto

import "bytes"

type InputFile struct {
	Name string
	Data bytes.Buffer
}

type Recognized struct {
	Text string `json:"text"`
}
