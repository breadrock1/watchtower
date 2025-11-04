package recognizer

import "bytes"

type RecognizeParams struct {
	FileName string
	FileData *bytes.Buffer
}
