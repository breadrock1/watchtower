package doc_storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"watchtower/internal/application/dto"
	"watchtower/internal/application/utils"
	"watchtower/internal/application/utils/telemetry"
)

const DocumentJsonMime = "application/json"
const ApiVersionPrefix = "/api/v1"

type DocSearcherClient struct {
	config *Config
}

func New(config *Config) *DocSearcherClient {
	return &DocSearcherClient{
		config: config,
	}
}

func (dsc *DocSearcherClient) StoreDocument(
	ctx context.Context,
	folder string,
	doc *dto.DocumentObject,
) (string, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "store-document")
	defer span.End()

	storeDoc := StoreDocumentForm{
		FileName:   doc.FileName,
		FilePath:   doc.FilePath,
		FileSize:   doc.FileSize,
		Content:    doc.Content,
		CreatedAt:  doc.CreatedAt.UnixMilli(),
		ModifiedAt: doc.ModifiedAt.UnixMilli(),
	}

	jsonData, err := json.Marshal(storeDoc)
	if err != nil {
		err = fmt.Errorf("failed while marshaling doc: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return "", err
	}

	buildURL := strings.Builder{}
	buildURL.WriteString(dsc.config.Address)
	buildURL.WriteString(fmt.Sprintf("%s/storage/%s/create?force=true", ApiVersionPrefix, folder))
	targetURL := buildURL.String()

	slog.Debug("storing document to index",
		slog.String("index", folder),
		slog.String("file-path", doc.FilePath),
	)

	reqBody := bytes.NewBuffer(jsonData)
	timeoutReq := time.Duration(300) * time.Second
	respData, err := utils.PUT(ctx, reqBody, targetURL, DocumentJsonMime, timeoutReq)
	if err != nil {
		err = fmt.Errorf("failed to store document to storage: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return "", err
	}

	status := &StoreDocumentResult{}
	err = json.Unmarshal(respData, status)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal response body: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return "", err
	}

	span.SetAttributes(
		attribute.String("index-id", folder),
		attribute.String("file-path", doc.FilePath),
	)

	return status.Message, nil
}

func (dsc *DocSearcherClient) UpdateDocument(_ context.Context, _ string, _ *dto.DocumentObject) error {
	return nil
}

func (dsc *DocSearcherClient) DeleteDocument(ctx context.Context, folder, id string) error {
	buildURL := strings.Builder{}
	buildURL.WriteString(dsc.config.Address)
	buildURL.WriteString(fmt.Sprintf("%s/storage/%s/%s", ApiVersionPrefix, folder, id))
	targetURL := buildURL.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, targetURL, bytes.NewReader([]byte{}))
	if err != nil {
		return fmt.Errorf("failed while creating new request: %w", err)
	}

	client := &http.Client{Timeout: time.Duration(100) * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed while sending request: %w", err)
	}

	if response.StatusCode/100 > 2 {
		return fmt.Errorf("bad response error status %s", response.Status)
	}

	return nil
}
