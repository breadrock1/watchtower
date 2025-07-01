package mocks

import (
	"context"
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

func (dsc *MockDocSearcherClient) StoreDocument(_ context.Context, _ string, doc *dto.StorageDocument) error {
	dsc.mu.Lock()
	defer dsc.mu.Unlock()
	dsc.storage[doc.ID] = doc
	return nil
}

func (dsc *MockDocSearcherClient) Get(id string) (*dto.StorageDocument, error) {
	return dsc.storage[id], nil
}

func (dsc *MockDocSearcherClient) GetDocuments() []*dto.StorageDocument {
	docs := make([]*dto.StorageDocument, len(dsc.storage))

	index := 0
	for _, val := range dsc.storage {
		docs[index] = val
		index++
	}
	return docs
}
