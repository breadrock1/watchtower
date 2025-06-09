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
)

const EmptyMessage = ""

type UseCase struct {
	watcherCh  <-chan dto.TaskEvent
	consumerCh <-chan dto.Message

	queue      task_queue.ITaskQueue
	cacher     task_manager.ITaskManager
	watcher    watcher.IWatcher
	recognizer recognizer.IRecognizer
	tokenizer  tokenizer.ITokenizer
	storage    doc_storage.IDocumentStorage
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
) *UseCase {
	return &UseCase{
		watcherCh:  watcherCh,
		consumerCh: consumerCh,
		queue:      queue,
		cacher:     cacher,
		watcher:    watcher,
		recognizer: recognizer,
		tokenizer:  tokenizer,
		storage:    storage,
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
					log.Printf("failed to pulish task to queue: %w", err)
				}

				uc.updateTaskStatus(ctx, &taskEvent, status, msg)
			case cMsg := <-uc.consumerCh:
				status, msg := dto.Successful, EmptyMessage
				if err := uc.processing(ctx, cMsg); err != nil {
					status, msg = dto.Failed, err.Error()
					log.Printf("failed while processing file: %w", err)
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

func (uc *UseCase) updateTaskStatus(ctx context.Context, taskEvent *dto.TaskEvent, status dto.TaskStatus, msg string) {
	taskEvent.Status, taskEvent.StatusText = status, msg
	if err := uc.cacher.Push(ctx, taskEvent); err != nil {
		log.Printf("failed to store task to cache: %w", err)
	}
}

func (uc *UseCase) processing(ctx context.Context, msg dto.Message) error {
	taskEvent := msg.Body
	uc.updateTaskStatus(ctx, &taskEvent, dto.Processing, EmptyMessage)

	fileData, err := uc.watcher.DownloadFile(ctx, taskEvent.Bucket, taskEvent.FilePath)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	inputFile := dto.InputFile{
		Name: path.Base(taskEvent.FilePath),
		Data: fileData,
	}
	recData, err := uc.recognizer.Recognize(inputFile)
	if err != nil {
		return fmt.Errorf("failed to recognize: %w", err)
	}

	var tokensRes *dto.Tokens
	tokensRes, err = uc.tokenizer.Load(recData.Text)
	if err != nil {
		log.Printf("failed to load tokens: %w", err)
	}

	doc := &dto.Document{
		FolderID:          taskEvent.Bucket,
		FolderPath:        path.Dir(taskEvent.FilePath),
		Content:           recData.Text,
		DocumentName:      path.Base(taskEvent.FilePath),
		DocumentPath:      taskEvent.FilePath,
		DocumentSize:      int64(fileData.Len()),
		DocumentExtension: path.Ext(taskEvent.FilePath),
		DocumentCreated:   taskEvent.CreatedAt,
		DocumentModified:  taskEvent.ModifiedAt,
		Tokens:            *tokensRes,
	}
	doc.ComputeMd5Hash()
	doc.ComputeSsdeepHash()

	if err = uc.storage.Store(doc); err != nil {
		return fmt.Errorf("failed to store doc %s: %w", doc.DocumentName, err)
	}

	return nil
}
