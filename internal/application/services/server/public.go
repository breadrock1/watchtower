package server

import "context"

const AppName = "watchtower"

type IServer interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}
