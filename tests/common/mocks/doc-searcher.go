package mocks

import (
	"sync"

	"watchtower/internal/application/dto"
)

type MockDocSearcherClient struct {
	mu      *sync.Mutex
	storage map[string]*dto.Document
}

func NewMockDocSearcherClient() *MockDocSearcherClient {
	return &MockDocSearcherClient{
		mu:      &sync.Mutex{},
		storage: make(map[string]*dto.Document),
	}
}

func (dsc *MockDocSearcherClient) Store(doc *dto.Document) error {
	dsc.mu.Lock()
	defer dsc.mu.Unlock()
	dsc.storage[doc.DocumentID] = doc
	return nil
}

func (dsc *MockDocSearcherClient) Get(id string) (*dto.Document, error) {
	return dsc.storage[id], nil
}

func (dsc *MockDocSearcherClient) GetDocuments() []*dto.Document {
	var docs []*dto.Document
	for _, val := range dsc.storage {
		docs = append(docs, val)
	}
	return docs
}
