package doc_storage

import (
	"context"

	"watchtower/internal/application/dto"
)

type IDocumentStorage interface {
	IIndexManager
	IDocumentManager
}

type IDocumentManager interface {
	DeleteDocument(ctx context.Context, folder, id string) error
	UpdateDocument(ctx context.Context, folder string, document *dto.DocumentObject) error
	StoreDocument(ctx context.Context, folder string, document *dto.DocumentObject) (string, error)
}

type IIndexManager interface {
	CreateIndex(ctx context.Context, folder string) error
	DeleteIndex(ctx context.Context, folder string) error
}
