package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonathanhecl/chunker"
	"watchtower/cmd"
	"watchtower/internal/application/usecase"
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
//	Processing -> 2;
//	Successful -> 3.
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

	settings := servConfig.Settings
	textChunker := chunker.NewChunker(
		settings.ChunkSize,
		settings.ChunkOverlap,
		chunker.DefaultSeparators,
		false,
		false,
	)

	cCtx, cancel := context.WithCancel(ctx)
	useCase := usecase.NewUseCase(
		*textChunker,
		rmqServ,
		redisServ,
		dedocServ,
		searcherServ,
		s3Serv,
	)
	useCase.LaunchWatcherListener(cCtx)

	httpServer := httpserver.New(&servConfig.Server, useCase)
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
