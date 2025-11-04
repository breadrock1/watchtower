package docstorage

import "context"

type IDocumentStorage interface {
	StoreDocument(ctx context.Context, document *Document) (DocumentID, error)
}
