package doc_storage

import (
	"context"

	"watchtower/internal/application/dto"
)

type IDocumentStorage interface {
	Store(ctx context.Context, folder string, document *dto.StorageDocument) error
}
