package doc_parser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"time"

	"watchtower/internal/application/models"
	"watchtower/internal/application/utils"
)

const RecognitionURL = "/parser/parse/text"

type DocParserClient struct {
	config *Config
}

func New(config *Config) *DocParserClient {
	return &DocParserClient{config}
}

func (dc *DocParserClient) Recognize(ctx context.Context, inputFile models.InputFile) (*models.Recognized, error) {
	var buf bytes.Buffer

	mpw := multipart.NewWriter(&buf)
	fileForm, err := mpw.CreateFormFile("file", inputFile.Name)
	if err != nil {
		err = fmt.Errorf("docparser: create file form error: %w", err)
		return nil, err
	}

	if _, err = fileForm.Write(inputFile.Data.Bytes()); err != nil {
		err = fmt.Errorf("docparser: write file form error: %w", err)
		return nil, err
	}

	if err = mpw.Close(); err != nil {
		return nil, err
	}

	mimeType := mpw.FormDataContentType()
	timeoutReq := dc.config.Timeout * time.Second
	targetURL := utils.BuildTargetURL(dc.config.Address, RecognitionURL)

	respData, err := utils.POST(ctx, &buf, targetURL, mimeType, timeoutReq)
	if err != nil {
		return nil, err
	}

	var recData Recognized
	_ = json.Unmarshal(respData, &recData)
	if len(recData.Text) == 0 {
		err = fmt.Errorf("docparser: returned empty content data")
		return nil, err
	}

	recognized := models.Recognized{Text: recData.Text}
	return &recognized, nil
}
