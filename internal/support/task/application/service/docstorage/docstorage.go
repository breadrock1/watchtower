package docstorage

import (
	"watchtower/internal/shared/kernel"
)

type IDocumentStorage interface {
	StoreDocument(ctx kernel.Ctx, document *Document) (DocumentID, error)
}
