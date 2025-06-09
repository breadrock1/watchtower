package server

import "context"

type IServer interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}
