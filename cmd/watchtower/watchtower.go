package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"watchtower/cmd"
	"watchtower/cmd/watchtower/httpserver"
	"watchtower/internal/core/cloud/infrastructure/s3"
	"watchtower/internal/process"
	"watchtower/internal/shared/telemetry"
	"watchtower/internal/support/task/infrastructure/docparser"
	"watchtower/internal/support/task/infrastructure/docsearch"
	"watchtower/internal/support/task/infrastructure/redis"
	"watchtower/internal/support/task/infrastructure/rmq"

	cloudApp "watchtower/internal/core/cloud/application"
	taskApp "watchtower/internal/support/task/application"
)

func main() {
	servConfig := cmd.Execute()

	traceProvider, err := telemetry.InitTracer(servConfig.Otlp.Tracer)
	if err != nil {
		slog.Warn("failed to init tracer", slog.String("err", err.Error()))
	}

	ctx := context.Background()

	taskStorage := redis.New(servConfig.Task.TaskStorage.Redis)
	taskQueue, err := rmq.New(servConfig.Task.TaskQueue.Rmq)
	if err != nil {
		log.Fatalf("task queue connection failed: %v", err)
	}
	err = taskQueue.StartConsuming(ctx)
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

	cCtx, cancel := context.WithCancel(ctx)
	storageUseCase := cloudApp.NewStorageUseCase(objStorage)
	taskUseCase := taskApp.NewTaskUseCase(taskStorage, taskQueue, docParser, docStorage)

	orchestrator := process.NewOrchestrator(servConfig.Orchestrator, storageUseCase, taskUseCase)
	orchestrator.LaunchListener(cCtx)

	httpServer := httpserver.SetupServer(servConfig.Otlp, orchestrator, traceProvider)
	go func() {
		if err := httpServer.Start(cCtx, servConfig.Server.Http); err != nil {
			slog.Error("http server start failed", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	cancel()
}
