package docparser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"
	"watchtower/internal/shared/metrics"

	"watchtower/internal/shared/kernel"
	"watchtower/internal/shared/utils"
	"watchtower/internal/support/task/application/service/recognizer"
)

const RecognitionURL = "/api/v1/parser/parse/text"

type DocParser struct {
	config Config
}

func (dc *DocParser) Health(ctx kernel.Ctx) error {
	targetURL := utils.BuildTargetURL(dc.config.Address, "/health")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return fmt.Errorf("docparser health check: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("docparser health check: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusInternalServerError {
		return fmt.Errorf("docparser health check: unexpected status %d", resp.StatusCode)
	}

	return nil
}

func New(config Config) recognizer.IRecognizer {
	return &DocParser{config}
}

func (dc *DocParser) Recognize(ctx kernel.Ctx, params *recognizer.RecognizeParams) (*recognizer.Recognized, error) {
	var buf bytes.Buffer

	mpw := multipart.NewWriter(&buf)
	fileForm, err := mpw.CreateFormFile("file", params.FileName)
	if err != nil {
		err = fmt.Errorf("docparser: create file form error: %w", err)
		return nil, err
	}

	if _, err = fileForm.Write(params.FileData.Bytes()); err != nil {
		err = fmt.Errorf("docparser: write file form error: %w", err)
		return nil, err
	}

	if err = mpw.Close(); err != nil {
		return nil, err
	}

	mimeType := mpw.FormDataContentType()
	timeoutReq := dc.config.Timeout * time.Second
	targetURL := utils.BuildTargetURL(dc.config.Address, RecognitionURL)

	start := time.Now()

	respData, err := utils.POST(ctx, &buf, targetURL, mimeType, timeoutReq)

	metrics.OutgoingHTTPRequestDurationSeconds.
		WithLabelValues(kernel.AppName, "recognize-document", "POST").
		Observe(time.Since(start).Seconds())

	if err != nil {
		return nil, err
	}

	var responseData ParsedContent
	_ = json.Unmarshal(respData, &responseData)
	if len(responseData.Text) == 0 {
		err = fmt.Errorf("docparser: returned empty content data")
		return nil, err
	}

	recData := responseData.ToRecognized()
	return &recData, nil
}
