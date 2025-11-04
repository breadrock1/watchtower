package docparser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"time"

	"watchtower/internal/shared/utils"
	"watchtower/internal/support/task/application/service/recognizer"
)

const RecognitionURL = "/parser/parse/text"

type DocParser struct {
	config *Config
}

func New(config *Config) *DocParser {
	return &DocParser{config}
}

func (dc *DocParser) Recognize(ctx context.Context, params recognizer.RecognizeParams) (recognizer.Recognized, error) {
	var buf bytes.Buffer
	var recData recognizer.Recognized

	mpw := multipart.NewWriter(&buf)
	fileForm, err := mpw.CreateFormFile("file", params.FileName)
	if err != nil {
		err = fmt.Errorf("docparser: create file form error: %w", err)
		return recData, err
	}

	if _, err = fileForm.Write(params.FileData.Bytes()); err != nil {
		err = fmt.Errorf("docparser: write file form error: %w", err)
		return recData, err
	}

	if err = mpw.Close(); err != nil {
		return recData, err
	}

	mimeType := mpw.FormDataContentType()
	timeoutReq := dc.config.Timeout * time.Second
	targetURL := utils.BuildTargetURL(dc.config.Address, RecognitionURL)

	respData, err := utils.POST(ctx, &buf, targetURL, mimeType, timeoutReq)
	if err != nil {
		return recData, err
	}

	var responseData ParsedContent
	_ = json.Unmarshal(respData, &responseData)
	if len(responseData.Text) == 0 {
		err = fmt.Errorf("docparser: returned empty content data")
		return recData, err
	}

	recData = responseData.ToRecognized()
	return recData, nil
}
