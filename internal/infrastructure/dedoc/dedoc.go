package dedoc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"time"

	"go.opentelemetry.io/otel/codes"
	"watchtower/internal/application/dto"
	"watchtower/internal/application/utils"
	"watchtower/internal/application/utils/telemetry"
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

func (dc *DedocClient) Recognize(ctx context.Context, inputFile dto.InputFile) (*dto.Recognized, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "recognize-file")
	defer span.End()

	var buf bytes.Buffer

	mpw := multipart.NewWriter(&buf)
	fileForm, err := mpw.CreateFormFile("file", inputFile.Name)
	if err != nil {
		err = fmt.Errorf("failed to create form file for dedoc: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if _, err = fileForm.Write(inputFile.Data.Bytes()); err != nil {
		err = fmt.Errorf("failed to write form file for dedoc: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if err = mpw.Close(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	mimeType := mpw.FormDataContentType()
	timeoutReq := dc.config.Timeout * time.Second
	targetURL := utils.BuildTargetURL(dc.config.Address, RecognitionURL)

	respData, err := utils.POST(ctx, &buf, targetURL, mimeType, timeoutReq)
	if err != nil {
		err = fmt.Errorf("failed send request: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var recData dto.Recognized
	_ = json.Unmarshal(respData, &recData)
	if len(recData.Text) == 0 {
		err = fmt.Errorf("returned empty content data")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &recData, nil
}
