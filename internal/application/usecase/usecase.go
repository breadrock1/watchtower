package usecase

import (
	"context"
	"fmt"
	"log"
	"path"

	"watchtower/internal/application/dto"
	"watchtower/internal/application/mapping"
	"watchtower/internal/application/services/doc-storage"
	"watchtower/internal/application/services/recognizer"
	"watchtower/internal/application/services/task-manager"
	"watchtower/internal/application/services/task-queue"
	"watchtower/internal/application/services/tokenizer"
	"watchtower/internal/application/services/watcher"
	"watchtower/internal/application/utils"
	"watchtower/internal/domain/core/structures"
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
				status, msg := domain.Pending, EmptyMessage
				if err := uc.publishToQueue(ctx, taskEvent); err != nil {
					status, msg = domain.Failed, err.Error()
					log.Printf("failed to pulish task to queue: %v", err)
				}

				uc.updateTaskStatus(ctx, &taskEvent, status, msg)
			case cMsg := <-uc.consumerCh:
				status, msg := domain.Successful, EmptyMessage
				if err := uc.processing(ctx, cMsg); err != nil {
					status, msg = domain.Failed, err.Error()
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

func (uc *UseCase) updateTaskStatus(ctx context.Context, task *dto.TaskEvent, status domain.TaskStatus, msg string) {
	task.Status, task.StatusText = mapping.TaskStatusToInt(status), msg
	if err := uc.cacher.Push(ctx, task); err != nil {
		log.Printf("failed to store task to cache: %v", err)
	}
}

func (uc *UseCase) processing(ctx context.Context, msg dto.Message) error {
	taskEvent := msg.Body
	uc.updateTaskStatus(ctx, &taskEvent, domain.Processing, EmptyMessage)

	fileData, err := uc.watcher.DownloadFile(ctx, taskEvent.Bucket, taskEvent.FilePath)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
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

	ssdeepHash, err := utils.ComputeSSDEEP(fileData.Bytes())
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

	if err = uc.storage.Store(ctx, taskEvent.Bucket, doc); err != nil {
		return fmt.Errorf("failed to store doc %s: %w", doc.FileName, err)
	}

	return nil
}
