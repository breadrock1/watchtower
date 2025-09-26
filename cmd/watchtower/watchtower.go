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
	"watchtower/internal/infrastructure/dedoc"
	"watchtower/internal/infrastructure/doc-storage"
	"watchtower/internal/infrastructure/httpserver"
	"watchtower/internal/infrastructure/redis"
	"watchtower/internal/infrastructure/rmq"
	"watchtower/internal/infrastructure/s3"
)

// @title          Watchtower service
// @version        0.0.1
// @description    Watchtower is a project designed to provide processing files created into cloud by events.
//
// @tag.name tasks
// @tag.description APIs to get status tasks. When TaskStatus may be:
// 	Failed -> -1;
//	Received -> 0;
//	Pending -> 1;
//	processConsumedTask -> 2;
//	Successful -> 3.
//
// @host      localhost:2893
// @BasePath  /api/v1
//
// @tag.name buckets
// @tag.description CRUD APIs to manage cloud buckets

// @tag.name files
// @tag.description CRUD APIs to manage files into bucket

// @tag.name share
// @tag.description Share files by URL API
func main() {
	ctx := context.Background()
	servConfig := cmd.Execute()

	dedocServ := dedoc.New(&servConfig.Ocr.Dedoc)
	redisServ := redis.New(&servConfig.Cacher.Redis)
	searcherServ := doc_storage.New(&servConfig.DocStorage.DocSearcher)

	rmqServ, err := rmq.New(&servConfig.Queue.Rmq)
	if err != nil {
		log.Fatalf("rmq connection failed: %v", err)
	}
	launchTasksConsumer(ctx, rmqServ)

	s3Serv, err := s3.New(&servConfig.Cloud.S3)
	if err != nil {
		log.Fatalf("s3 connection failed: %v", err)
	}

	cCtx, cancel := context.WithCancel(ctx)
	taskMangerUC := usecase.NewTaskManagerUseCase(redisServ)
	storageUC := usecase.NewStorageUseCase(searcherServ, s3Serv)
	processorUC := usecase.NewPipelineUseCase(storageUC, taskMangerUC, rmqServ, dedocServ)
	processorUC.LaunchWatcherListener(cCtx)

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

func launchTasksConsumer(ctx context.Context, rmqServ *rmq.RmqClient) {
	if err := rmqServ.Consume(ctx); err != nil {
		log.Fatalf("rmq consumer launching failed: %v", err)
	}
}
