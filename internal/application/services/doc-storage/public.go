package doc_storage

import (
	"watchtower/internal/application/dto"
)

type IDocumentStorage interface {
	Store(document *dto.Document) error
}
