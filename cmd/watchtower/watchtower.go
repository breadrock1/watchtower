package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"watchtower/cmd"
	"watchtower/internal/application/dto"
	"watchtower/internal/application/usecase"
	"watchtower/internal/infrastructure/dedoc"
	"watchtower/internal/infrastructure/doc-searcher"
	"watchtower/internal/infrastructure/httpserver"
	"watchtower/internal/infrastructure/pg"
	"watchtower/internal/infrastructure/redis"
	"watchtower/internal/infrastructure/rmq"
	"watchtower/internal/infrastructure/s3"
	"watchtower/internal/infrastructure/vectorizer"
)

// @title          Watchtower service
// @version        0.0.1
// @description    Watchtower is a project designed to provide processing files created into cloud by events.
//
// @tag.name watcher
// @tag.description APIs to manage cloud watchers
//
// @tag.name tasks
// @tag.description APIs to get status tasks

func main() {
	ctx := context.Background()
	servConfig := cmd.Execute()

	dedocServ := dedoc.New(&servConfig.Ocr.Dedoc)
	redisServ := redis.New(&servConfig.Cacher.Redis)
	searcherServ := doc_searcher.New(&servConfig.Storage.Docsearcher)
	vectorizerServ := vectorizer.New(&servConfig.Tokenizer.Vectorizer)

	rmqServ, err := rmq.New(&servConfig.Queue.Rmq)
	if err != nil {
		log.Fatalf("rmq connection failed: %v", err)
	}
	launchTasksConsumer(ctx, rmqServ)

	s3Serv, err := s3.New(&servConfig.Cloud.Minio)
	if err != nil {
		log.Fatalf("s3 connection failed: %v", err)
	}
	launchBucketListeners(ctx, s3Serv, servConfig.Cloud.Minio.WatchedDirs)

	pgStorage, err := pg.NewPgClient(&servConfig.Watcher.Storage.Pg)
	if err != nil {
		log.Printf("failed to connect to pg: %v", err)
	}

	cCtx, cancel := context.WithCancel(ctx)
	useCase := usecase.New(
		s3Serv.GetEventsChannel(),
		rmqServ.GetConsumerChannel(),
		rmqServ,
		redisServ,
		s3Serv,
		dedocServ,
		vectorizerServ,
		searcherServ,
		pgStorage,
	)
	useCase.LaunchProcessing(cCtx)

	if pgStorage != nil {
		useCase.LoadAndLaunchWatchedDirs(cCtx)
	}

	httpServer := httpserver.New(&servConfig.Server.Http, redisServ, s3Serv)
	go func() {
		if err := httpServer.Start(ctx); err != nil {
			log.Fatalf("http server start failed: %v", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-ch
	cancel()

	if pgStorage != nil {
		useCase.StoreWatchedDirs(cCtx)
	}
}

func launchBucketListeners(ctx context.Context, s3Serv *s3.S3Client, dirs []string) {
	watchDirs := make([]dto.Directory, len(dirs))
	for index, dirName := range dirs {
		watchDirs[index] = dto.FromBucketName(dirName)
	}

	if err := s3Serv.LaunchWatcher(ctx, watchDirs); err != nil {
		log.Printf("failed to start s3 bucket listers: %v", err)
	}
}

func launchTasksConsumer(ctx context.Context, rmqServ *rmq.RmqClient) {
	if err := rmqServ.Consume(ctx); err != nil {
		log.Fatalf("rmq consumer launching failed: %v", err)
	}
}
