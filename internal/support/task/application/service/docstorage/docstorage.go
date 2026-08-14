package docstorage

import (
	"watchtower/internal/shared/kernel"
)

type IDocumentStorage interface {
	kernel.IHealth

	StoreDocument(ctx kernel.Ctx, document *Document) (DocumentID, error)
}
