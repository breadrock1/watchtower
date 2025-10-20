package docstorage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"watchtower/internal/application/utils"
	"watchtower/internal/domain/core/structures"
)

const DocumentJsonMime = "application/json"

type DocSearchClient struct {
	config *Config
}

func New(config *Config) *DocSearchClient {
	return &DocSearchClient{
		config: config,
	}
}

func (d *DocSearchClient) StoreDocument(ctx context.Context, doc *domain.Document) (string, error) {
	index := doc.Index
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
		err = fmt.Errorf("serialize error: %w", err)
		return "", err
	}

	buildURL := strings.Builder{}
	buildURL.WriteString(d.config.Address)
	buildURL.WriteString(fmt.Sprintf("/storage/%s/create?force=true", index))
	targetURL := buildURL.String()

	slog.Debug("storing document to index",
		slog.String("index", index),
		slog.String("file-path", doc.FilePath),
	)

	reqBody := bytes.NewBuffer(jsonData)
	timeoutReq := d.config.Timeout * time.Second
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
