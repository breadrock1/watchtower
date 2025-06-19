package doc_storage

import "watchtower/internal/application/dto"

type IDocumentStorage interface {
	Store(folder string, document *dto.StorageDocument) error
}
