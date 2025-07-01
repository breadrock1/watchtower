package mocks

import (
	"context"
	"time"

	"watchtower/internal/application/dto"

	_ "github.com/lib/pq"
)

type MockPgClient struct{}

func (pc *MockPgClient) LoadAllWatcherDirs(_ context.Context) ([]dto.Directory, error) {
	dir := dto.Directory{
		"test-bucket",
		"./",
		time.Now(),
	}

	return []dto.Directory{dir}, nil
}

func (pc *MockPgClient) StoreWatcherDir(_ context.Context, _ dto.Directory) error {
	return nil
}

func (pc *MockPgClient) DeleteWatcherDir(_ context.Context, _, _ string) error {
	return nil
}

func (pc *MockPgClient) DeleteAllWatcherDirs(_ context.Context) error {
	return nil
}
