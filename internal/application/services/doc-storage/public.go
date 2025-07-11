package doc_storage

import (
	"context"

	"watchtower/internal/application/dto"
)

type IDocumentStorage interface {
	StoreDocument(ctx context.Context, folder string, document *dto.StorageDocument) (string, error)
	UpdateDocument(ctx context.Context, folder string, document *dto.StorageDocument) error
	DeleteDocument(ctx context.Context, folder, id string) error
	CreateIndex(ctx context.Context, folder string) error
	DeleteIndex(ctx context.Context, folder string) error
}
