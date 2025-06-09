package dedoc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"time"

	"watchtower/internal/application/dto"
	"watchtower/internal/application/utils"
)

const RecognitionURL = "/ocr_extract_text"

type DedocClient struct {
	config *Config
}

func New(config *Config) *DedocClient {
	return &DedocClient{
		config: config,
	}
}

func (dc *DedocClient) Recognize(inputFile dto.InputFile) (*dto.Recognized, error) {
	var buf bytes.Buffer

	mpw := multipart.NewWriter(&buf)
	fileForm, err := mpw.CreateFormFile("file", inputFile.Name)
	if err != nil {
		return nil, err
	}

	if _, err = fileForm.Write(inputFile.Data.Bytes()); err != nil {
		return nil, err
	}

	if err = mpw.Close(); err != nil {
		return nil, err
	}

	mimeType := mpw.FormDataContentType()
	timeoutReq := dc.config.Timeout * time.Second
	targetURL := utils.BuildTargetURL(dc.config.EnableSSL, dc.config.Address, RecognitionURL)

	respData, err := utils.POST(&buf, targetURL, mimeType, timeoutReq)
	if err != nil {
		return nil, fmt.Errorf("failed send request: %v", err)
	}

	var recData dto.Recognized
	_ = json.Unmarshal(respData, &recData)
	if len(recData.Text) == 0 {
		return nil, fmt.Errorf("returned empty content data")
	}

	return &recData, nil
}
