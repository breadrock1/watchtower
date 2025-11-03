package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"watchtower/cmd"
	"watchtower/internal/application/service/server"
	"watchtower/internal/application/usecase"
	"watchtower/internal/application/utils/telemetry"
	"watchtower/internal/infrastructure/docparser"
	"watchtower/internal/infrastructure/docsearch"
	"watchtower/internal/infrastructure/httpserver"
	"watchtower/internal/infrastructure/redis"
	"watchtower/internal/infrastructure/rmq"
	"watchtower/internal/infrastructure/s3"
)

func main() {
	ctx := context.Background()
	servConfig := cmd.Execute()

	recognizer := docparser.New(&servConfig.Recognizer.DocParser)

	taskStorage := redis.New(&servConfig.Task.TaskStorage.Redis)
	taskQueue, err := rmq.New(&servConfig.Task.TaskQueue.Rmq)
	if err != nil {
		log.Fatalf("task queue connection failed: %v", err)
	}
	launchTasksConsumer(ctx, taskQueue)

	docStorage := docsearch.New(&servConfig.Storage.DocumentStorage.DocSearcher)
	objStorage, err := s3.New(&servConfig.Storage.ObjectStorage.S3)
	if err != nil {
		log.Fatalf("object storage connection failed: %v", err)
	}

	cCtx, cancel := context.WithCancel(ctx)
	storageUseCase := usecase.NewStorageUseCase(objStorage)
	processingUseCase := usecase.NewProcessingUseCase(objStorage, taskStorage, taskQueue, recognizer, docStorage)
	processingUseCase.LaunchListener(cCtx)

	traceProvider, err := telemetry.InitTracer(servConfig.Server.Tracer)
	if err != nil {
		slog.Warn("failed to init tracer", slog.String("err", err.Error()))
	}

	serverState := server.New(storageUseCase, processingUseCase)
	httpServer := httpserver.New(&servConfig.Server, serverState, traceProvider)
	go func() {
		if err := httpServer.Start(cCtx); err != nil {
			log.Fatalf("http server start failed: %v", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	cancel()
}

func launchTasksConsumer(ctx context.Context, rmqServ *rmq.RabbitMQClient) {
	if err := rmqServ.StartConsuming(ctx); err != nil {
		log.Fatalf("failed to launch task queue consumer: %v", err)
	}
}
