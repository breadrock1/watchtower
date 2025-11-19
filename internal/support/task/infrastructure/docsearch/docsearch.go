package docsearch

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/enetx/g"
	"github.com/enetx/surf"

	"watchtower/internal/support/task/application/service/docstorage"
)

type DocSearch struct {
	config Config
}

func New(config Config) docstorage.IDocumentStorage {
	return &DocSearch{
		config: config,
	}
}

func (ds *DocSearch) StoreDocument(ctx context.Context, doc *docstorage.Document) (docstorage.DocumentID, error) {
	index := doc.Index
	storeDoc := StoreDocumentForm{
		FileName:   doc.Name,
		FilePath:   doc.Path,
		FileSize:   doc.Size,
		Content:    doc.Content,
		CreatedAt:  doc.CreatedAt.UnixMilli(),
		ModifiedAt: doc.ModifiedAt.UnixMilli(),
	}

	buildURL := strings.Builder{}
	buildURL.WriteString(ds.config.Address)
	buildURL.WriteString(fmt.Sprintf("/storage/%s/create?force=true", index))
	targetURL := buildURL.String()

	slog.Debug("storing document to index",
		slog.String("index", index),
		slog.String("file-path", doc.Path),
	)

	resp := surf.NewClient().
		Post(g.String(targetURL), storeDoc).
		WithContext(ctx).
		Do()

	if resp.IsErr() {
		err := fmt.Errorf("http-request error: %w", resp.Err())
		return "", err
	}

	var storeResult StoreDocumentResult
	resp.Ok().Body.JSON(&storeResult)
	return storeResult.Message, nil
}
