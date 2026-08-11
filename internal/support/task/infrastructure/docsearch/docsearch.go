package docsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"watchtower/internal/shared/kernel"
	"watchtower/internal/shared/utils"
	"watchtower/internal/support/task/application/service/docstorage"
)

const DocumentJsonMime = "application/json"

type DocSearch struct {
	config Config
}

func (ds *DocSearch) Health(ctx kernel.Ctx) error {
	targetURL := ds.config.Address + "/health"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return fmt.Errorf("docsearch health check: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("docsearch health check: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusInternalServerError {
		return fmt.Errorf("docsearch health check: unexpected status %d", resp.StatusCode)
	}

	return nil
}

func New(config Config) docstorage.IDocumentStorage {
	return &DocSearch{
		config: config,
	}
}

func (ds *DocSearch) StoreDocument(ctx kernel.Ctx, doc *docstorage.Document) (docstorage.DocumentID, error) {
	index := doc.Index
	storeDoc := StoreDocumentForm{
		FileName:   doc.Name,
		FilePath:   doc.Path,
		FileSize:   doc.Size,
		Content:    doc.Content,
		CreatedAt:  doc.CreatedAt.UnixMilli(),
		ModifiedAt: doc.ModifiedAt.UnixMilli(),
	}

	jsonData, err := json.Marshal(storeDoc)
	if err != nil {
		err = fmt.Errorf("serialize error: %w", err)
		return "", err
	}

	urlPath := fmt.Sprintf("/api/v1/storage/%s/create?force=true", index)
	buildURL := &strings.Builder{}
	buildURL.WriteString(ds.config.Address)
	buildURL.WriteString(urlPath)
	targetURL := buildURL.String()

	slog.Debug("storing document to index",
		slog.String("index", index),
		slog.String("file-path", doc.Path),
	)

	reqBody := bytes.NewBuffer(jsonData)
	timeoutReq := ds.config.Timeout * time.Second
	respData, err := utils.PUT(ctx, reqBody, targetURL, DocumentJsonMime, timeoutReq)
	if err != nil {
		err = fmt.Errorf("http-request error: %w", err)
		return "", err
	}

	status := &StoreDocumentResult{}
	err = json.Unmarshal(respData, status)
	if err != nil {
		err = fmt.Errorf("deserialize error: %w", err)
		return "", err
	}

	return status.Message, nil
}
