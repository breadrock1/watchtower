package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/breadrock1/otlp-go/otlp"

	"watchtower/cmd"
	"watchtower/cmd/watchtower/httpserver"
	"watchtower/internal/core/cloud/infrastructure/s3"
	"watchtower/internal/process"
	"watchtower/internal/support/task/infrastructure/docparser"
	"watchtower/internal/support/task/infrastructure/docsearch"
	"watchtower/internal/support/task/infrastructure/redis"
	"watchtower/internal/support/task/infrastructure/rmq"

	cloudApp "watchtower/internal/core/cloud/application"
	taskApp "watchtower/internal/support/task/application"
)

const (
	ShutdownDuration = 10 * time.Second
)

func main() {
	servConfig := cmd.Execute()

	logger := otlp_go.InitLocalLogger(servConfig.Otlp)
	slog.SetDefault(logger)

	ctx := context.Background()
	cCtx, cancel := context.WithCancel(ctx)

	taskStorage := redis.New(servConfig.Task.TaskStorage.Redis)
	taskQueue, err := rmq.New(servConfig.Task.TaskQueue.Rmq)
	if err != nil {
		slog.Error("task queue connection failed", slog.String("err", err.Error()))
		os.Exit(1)
	}
	err = taskQueue.StartConsuming(cCtx)
	if err != nil {
		slog.Error("failed to launch task queue consumer", slog.String("err", err.Error()))
		os.Exit(1)
	}

	docParser := docparser.New(servConfig.Task.Processor.DocParser)
	docStorage := docsearch.New(servConfig.Task.Processor.DocStorage)
	objStorage, err := s3.New(servConfig.Storage.S3)
	if err != nil {
		slog.Error("object storage connection failed", slog.String("err", err.Error()))
		os.Exit(1)
	}

	storageUseCase := cloudApp.NewStorageUseCase(objStorage)
	taskUseCase := taskApp.NewTaskUseCase(taskStorage, taskQueue, docParser, docStorage)

	orchestrator := process.NewOrchestrator(servConfig.Orchestrator, storageUseCase, taskUseCase)
	orchestrator.LaunchListener(cCtx)

	httpServer := httpserver.SetupServer(servConfig.Otlp, orchestrator)
	go func() {
		if err := httpServer.Start(servConfig.Server.Http); err != nil {
			slog.Error("http server start failed", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	slog.Warn("received shutdown signal. shutdown server...")

	cancel()

	shutdownCtx, shutdownRelease := context.WithTimeout(ctx, ShutdownDuration)
	defer shutdownRelease()

	if err = httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("http server shutdown failed", slog.String("err", err.Error()))
		return
	}

	time.Sleep(time.Second)

	slog.Info("application has been shutdown successfully")
}
