package mocks

import (
	"context"
	"sync"

	"github.com/google/uuid"
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

func (dsc *MockDocSearcherClient) StoreDocument(_ context.Context, _ string, doc *dto.StorageDocument) (string, error) {
	id := uuid.New().String()
	dsc.mu.Lock()
	defer dsc.mu.Unlock()
	dsc.storage[id] = doc
	return id, nil
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

func (dsc *MockDocSearcherClient) UpdateDocument(_ context.Context, _ string, _ *dto.StorageDocument) error {
	return nil
}

func (dsc *MockDocSearcherClient) DeleteDocument(_ context.Context, _, _ string) error {
	return nil
}

func (dsc *MockDocSearcherClient) CreateIndex(_ context.Context, _ string) error {
	return nil
}

func (dsc *MockDocSearcherClient) DeleteIndex(_ context.Context, _ string) error {
	return nil
}
