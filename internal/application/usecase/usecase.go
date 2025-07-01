package usecase

import (
	"context"
	"fmt"
	"log"
	"path"

	"watchtower/internal/application/dto"
	"watchtower/internal/application/services/doc-storage"
	"watchtower/internal/application/services/recognizer"
	"watchtower/internal/application/services/task-manager"
	"watchtower/internal/application/services/task-queue"
	"watchtower/internal/application/services/tokenizer"
	"watchtower/internal/application/services/watcher"
	"watchtower/internal/application/utils"
)

const EmptyMessage = ""

type UseCase struct {
	watcherCh  <-chan dto.TaskEvent
	consumerCh <-chan dto.Message

	queue        task_queue.ITaskQueue
	cacher       task_manager.ITaskManager
	watcher      watcher.IWatcher
	recognizer   recognizer.IRecognizer
	tokenizer    tokenizer.ITokenizer
	storage      doc_storage.IDocumentStorage
	watchStorage watcher.IConfigStorage
}

func New(
	watcherCh chan dto.TaskEvent,
	consumerCh chan dto.Message,
	queue task_queue.ITaskQueue,
	cacher task_manager.ITaskManager,
	watcher watcher.IWatcher,
	recognizer recognizer.IRecognizer,
	tokenizer tokenizer.ITokenizer,
	storage doc_storage.IDocumentStorage,
	watchStorage watcher.IConfigStorage,
) *UseCase {
	return &UseCase{
		watcherCh:    watcherCh,
		consumerCh:   consumerCh,
		queue:        queue,
		cacher:       cacher,
		watcher:      watcher,
		recognizer:   recognizer,
		tokenizer:    tokenizer,
		storage:      storage,
		watchStorage: watchStorage,
	}
}

func (uc *UseCase) LaunchProcessing(ctx context.Context) {
	go func() {
		for {
			select {
			case taskEvent := <-uc.watcherCh:
				status, msg := dto.Pending, EmptyMessage
				if err := uc.publishToQueue(ctx, taskEvent); err != nil {
					status, msg = dto.Failed, err.Error()
					log.Printf("failed to pulish task to queue: %v", err)
				}

				uc.updateTaskStatus(ctx, &taskEvent, status, msg)
			case cMsg := <-uc.consumerCh:
				status, msg := dto.Successful, EmptyMessage
				if err := uc.Processing(ctx, cMsg); err != nil {
					status, msg = dto.Failed, err.Error()
					log.Printf("failed while processing file: %v", err)
				}

				uc.updateTaskStatus(ctx, &cMsg.Body, status, msg)
			case <-ctx.Done():
				log.Println("terminated processing")
				return
			}
		}
	}()
}

func (uc *UseCase) publishToQueue(ctx context.Context, taskEvent dto.TaskEvent) error {
	msg := dto.FromTaskEvent(taskEvent)
	return uc.queue.Publish(ctx, msg)
}

func (uc *UseCase) updateTaskStatus(ctx context.Context, task *dto.TaskEvent, status dto.TaskStatus, msg string) {
	task.Status, task.StatusText = status, msg
	if err := uc.cacher.Push(ctx, task); err != nil {
		log.Printf("failed to store task to cache: %v", err)
	}
}

func (uc *UseCase) Processing(ctx context.Context, msg dto.Message) error {
	taskEvent := msg.Body
	uc.updateTaskStatus(ctx, &taskEvent, dto.Processing, EmptyMessage)

	var callback func(ctx context.Context, task dto.TaskEvent) error
	switch taskEvent.EventType {
	case dto.CreateFile:
		callback = uc.processFile

	case dto.CreateBucket:
		callback = uc.createBucket

	case dto.DeleteBucket:
		callback = uc.deleteBucket
	default:
		return fmt.Errorf("unknown task event type: %v", taskEvent.EventType)
	}

	err := callback(ctx, taskEvent)
	return err
}

func (uc *UseCase) createBucket(ctx context.Context, taskEvent dto.TaskEvent) error {
	err := uc.storage.CreateIndex(ctx, taskEvent.Bucket)
	if err != nil {
		return fmt.Errorf("failed to create index: %v", err)
	}
	return nil
}

func (uc *UseCase) deleteBucket(ctx context.Context, taskEvent dto.TaskEvent) error {
	err := uc.storage.DeleteIndex(ctx, taskEvent.Bucket)
	if err != nil {
		return fmt.Errorf("failed to delete index: %v", err)
	}
	return nil
}

func (uc *UseCase) processFile(ctx context.Context, taskEvent dto.TaskEvent) error {
	fileData, err := uc.watcher.DownloadFile(ctx, taskEvent.Bucket, taskEvent.FilePath)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	ssdeepHash, err := utils.ComputeSSDEEP(fileData.Bytes())
	if err != nil {
		log.Printf("failed to compute SSDEEP hash: %v", err)
	}

	inputFile := dto.InputFile{
		Name: path.Base(taskEvent.FilePath),
		Data: fileData,
	}
	recData, err := uc.recognizer.Recognize(ctx, inputFile)
	if err != nil {
		return fmt.Errorf("failed to recognize: %w", err)
	}

	var tokensRes *dto.ComputedTokens
	tokensRes, err = uc.tokenizer.Load(ctx, recData.Text)
	if err != nil {
		log.Printf("failed to load tokens: %v", err)
	}

	doc := &dto.StorageDocument{
		Content:    recData.Text,
		SSDEEP:     ssdeepHash,
		ID:         utils.ComputeMd5(fileData.Bytes()),
		Class:      "unknown",
		FileName:   path.Base(taskEvent.FilePath),
		FilePath:   taskEvent.FilePath,
		FileSize:   fileData.Len(),
		CreatedAt:  taskEvent.CreatedAt,
		ModifiedAt: taskEvent.ModifiedAt,
		Tokens:     *tokensRes,
	}

	if err = uc.storage.StoreDocument(ctx, taskEvent.Bucket, doc); err != nil {
		return fmt.Errorf("failed to store doc %s: %w", doc.FileName, err)
	}

	return nil
}

func (uc *UseCase) LoadAndLaunchWatchedDirs(ctx context.Context) {
	dirs, err := uc.watchStorage.LoadAllWatcherDirs(ctx)
	if err != nil {
		log.Printf("failed to load all watcher dirs: %v", err)
	}

	for _, dir := range dirs {
		err = uc.watcher.AttachWatchedDir(ctx, dir)
		if err != nil {
			log.Printf("failed to attach watcher dir %s: %v", dir, err)
		}
	}
}

func (uc *UseCase) StoreWatchedDirs(ctx context.Context) {
	dirs, err := uc.watcher.GetWatchedDirs(ctx)
	if err != nil {
		log.Printf("failed to get all watcher dirs: %v", err)
	}

	for _, dir := range dirs {
		err = uc.watchStorage.StoreWatcherDir(ctx, dir)
		if err != nil {
			log.Printf("failed to store watcher dir %s: %v", dir, err)
		}
	}
}
