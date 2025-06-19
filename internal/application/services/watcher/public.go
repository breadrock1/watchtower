package watcher

import (
	"bytes"
	"context"

	"watchtower/internal/application/dto"
)

type IWatcher interface {
	IFileStorage
	IWatchManager

	LaunchWatcher(ctx context.Context, dirs []dto.Directory) error
	TerminateWatcher(ctx context.Context) error
}

type IWatchManager interface {
	GetWatchedDirs(ctx context.Context) ([]dto.Directory, error)
	AttachWatchedDir(ctx context.Context, dir dto.Directory) error
	DetachWatchedDir(ctx context.Context, path string) error
}

type IFileStorage interface {
	UploadFile(ctx context.Context, bucket, filePath string, data *bytes.Buffer) error
	DownloadFile(ctx context.Context, bucket, filePath string) (bytes.Buffer, error)
}
