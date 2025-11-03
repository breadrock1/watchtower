package docsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"watchtower/internal/application/utils"
	"watchtower/internal/domain/support/docstorage"
)

const DocumentJsonMime = "application/json"

type DocSearch struct {
	config *Config
}

func New(config *Config) *DocSearch {
	return &DocSearch{
		config: config,
	}
}

func (ds *DocSearch) StoreDocument(ctx context.Context, doc docstorage.Document) (docstorage.DocumentID, error) {
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

	buildURL := strings.Builder{}
	buildURL.WriteString(ds.config.Address)
	buildURL.WriteString(fmt.Sprintf("/storage/%s/create?force=true", index))
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
