package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"watchtower/cmd"
	"watchtower/internal/application/usecase"
	"watchtower/internal/application/utils/telemetry"
	"watchtower/internal/infrastructure/doc-parser"
	"watchtower/internal/infrastructure/doc-storage"
	"watchtower/internal/infrastructure/httpserver"
	"watchtower/internal/infrastructure/redis"
	"watchtower/internal/infrastructure/rmq"
	"watchtower/internal/infrastructure/s3"
)

func main() {
	ctx := context.Background()
	servConfig := cmd.Execute()

	recognizer := doc_parser.New(&servConfig.Recognizer.DocParser)

	taskStorage := redis.New(&servConfig.Task.TaskStorage.Redis)
	taskQueue, err := rmq.New(&servConfig.Task.TaskQueue.Rmq)
	if err != nil {
		log.Fatalf("rmq connection failed: %v", err)
	}
	launchTasksConsumer(ctx, taskQueue)

	docStorage := doc_storage.New(&servConfig.Storage.DocumentStorage.DocSearcher)
	objStorage, err := s3.New(&servConfig.Storage.ObjectStorage.S3)
	if err != nil {
		log.Fatalf("s3 connection failed: %v", err)
	}

	cCtx, cancel := context.WithCancel(ctx)
	taskMangerUC := usecase.NewTaskUseCase(taskStorage, taskQueue)
	storageUC := usecase.NewStorageUseCase(docStorage, objStorage)
	processorUC := usecase.NewPipelineUseCase(storageUC, taskMangerUC, recognizer)
	processorUC.LaunchListener(cCtx)

	traceProvider, err := telemetry.InitTracer(servConfig.Server.Tracer)
	if err != nil {
		slog.Warn("failed to init tracer", slog.String("err", err.Error()))
	}

	httpServer := httpserver.New(&servConfig.Server, traceProvider, processorUC, storageUC, taskMangerUC)
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
	if err := rmqServ.Consume(ctx); err != nil {
		log.Fatalf("failed to launch task queue consumer: %v", err)
	}
}
