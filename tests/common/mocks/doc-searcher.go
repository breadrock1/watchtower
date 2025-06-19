package mocks

import (
	"sync"

	"watchtower/internal/application/dto"
)

type MockDocSearcherClient struct {
	mu      *sync.Mutex
	storage map[string]*dto.StorageDocument
}

func NewMockDocSearcherClient() *MockDocSearcherClient {
	return &MockDocSearcherClient{
		mu:      &sync.Mutex{},
		storage: make(map[string]*dto.StorageDocument),
	}
}

func (dsc *MockDocSearcherClient) Store(_ string, doc *dto.StorageDocument) error {
	dsc.mu.Lock()
	defer dsc.mu.Unlock()
	dsc.storage[doc.ID] = doc
	return nil
}

func (dsc *MockDocSearcherClient) Get(id string) (*dto.StorageDocument, error) {
	return dsc.storage[id], nil
}

func (dsc *MockDocSearcherClient) GetDocuments() []*dto.StorageDocument {
	var docs []*dto.StorageDocument
	for _, val := range dsc.storage {
		docs = append(docs, val)
	}
	return docs
}
