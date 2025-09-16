package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"sync"
	"time"

	"github.com/jonathanhecl/chunker"
	"golang.org/x/sync/semaphore"
	"watchtower/internal/application/dto"
	"watchtower/internal/application/mapping"
	"watchtower/internal/application/services/doc-storage"
	"watchtower/internal/application/services/object-storage"
	"watchtower/internal/application/services/recognizer"
	"watchtower/internal/application/services/task-manager"
	"watchtower/internal/application/services/task-queue"
	"watchtower/internal/application/utils"
)

const EmptyMessage = ""
const SEMAPHORE_WORKERS_COUNT = 10

type UseCase struct {
	processorCh chan dto.TaskEvent
	consumerCh  <-chan dto.Message
	textChunker chunker.Chunker

	taskQueue   task_queue.ITaskQueue
	taskManager task_manager.ITaskManager
	recognizer  recognizer.IRecognizer
	docStorage  doc_storage.IDocumentStorage
	objStorage  object_storage.IObjectStorage
}

func NewUseCase(
	textChunker chunker.Chunker,
	taskQueue task_queue.ITaskQueue,
	taskManager task_manager.ITaskManager,
	recognizer recognizer.IRecognizer,
	docStorage doc_storage.IDocumentStorage,
	objStorage object_storage.IObjectStorage,
) *UseCase {
	consumerCh := taskQueue.GetConsumerChannel()
	processorCh := make(chan dto.TaskEvent)
	return &UseCase{
		textChunker: textChunker,
		processorCh: processorCh,
		consumerCh:  consumerCh,
		taskQueue:   taskQueue,
		taskManager: taskManager,
		recognizer:  recognizer,
		docStorage:  docStorage,
		objStorage:  objStorage,
	}
}

func (uc *UseCase) LaunchWatcherListener(ctx context.Context) {
	go func() {
		for {
			select {
			case taskEvent := <-uc.processorCh:
				// TODO: Disabled for TechDebt
				_ = uc.isTaskAlreadyProcessed(ctx, &taskEvent)
				// if uc.isTaskAlreadyProcessed(ctx, &taskEvent) {
				//	 log.Printf("task has been already processed: %s", taskEvent.ID)
				//	 continue
				// }

				status, msg := dto.Pending, EmptyMessage
				if err := uc.publishToQueue(ctx, taskEvent); err != nil {
					status, msg = dto.Failed, err.Error()
					slog.Error("failed to publish task to queue",
						slog.String("task-id", taskEvent.ID),
						slog.String("err", err.Error()))
				}

				uc.updateTaskStatus(ctx, &taskEvent, status, msg)
			case cMsg := <-uc.consumerCh:
				uc.Processing(ctx, cMsg)
			case <-ctx.Done():
				slog.Info("terminated file processing")
				return
			}
		}
	}()
}

func (uc *UseCase) Processing(ctx context.Context, recvMsg dto.Message) {
	taskEvent := recvMsg.Body
	slog.Info("processing task event",
		slog.String("task-id", taskEvent.ID),
		slog.String("bucket", taskEvent.Bucket),
		slog.String("file-path", taskEvent.FilePath),
		slog.Time("created", taskEvent.CreatedAt))

	uc.updateTaskStatus(ctx, &taskEvent, dto.Processing, EmptyMessage)

	status := dto.Successful
	msg, err := uc.processFile(ctx, recvMsg.Body)
	if err != nil {
		status, msg = dto.Failed, err.Error()
		slog.Error("failed while task processing",
			slog.String("task-id", taskEvent.ID),
			slog.String("err", err.Error()))
	}

	taskEvent.StatusText = msg
	uc.updateTaskStatus(ctx, &taskEvent, status, msg)
}

func (uc *UseCase) publishToQueue(ctx context.Context, taskEvent dto.TaskEvent) error {
	msg := mapping.MessageFromTaskEvent(taskEvent)
	err := uc.taskQueue.Publish(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to publish task event to queue: %w", err)
	}
	return nil
}

func (uc *UseCase) updateTaskStatus(ctx context.Context, task *dto.TaskEvent, status dto.TaskStatus, msg string) {
	task.Status, task.StatusText = status, msg
	if err := uc.taskManager.Push(ctx, task); err != nil {
		slog.Error("failed caching task status",
			slog.String("task-id", task.ID),
			slog.String("err", err.Error()))
	}
}

func (uc *UseCase) isTaskAlreadyProcessed(ctx context.Context, task *dto.TaskEvent) bool {
	storageTask, err := uc.taskManager.Get(ctx, task.Bucket, task.ID)
	if err != nil {
		slog.Error("failed to get task from cache",
			slog.String("task-id", task.ID),
			slog.String("err", err.Error()))
		return false
	}

	if storageTask == nil {
		return false
	}

	switch storageTask.Status {
	case dto.Received:
		fallthrough
	case dto.Pending:
		fallthrough
	case dto.Processing:
		return true
	case dto.Failed:
		fallthrough
	case dto.Successful:
		return false
	default:
		return false
	}
}

func (uc *UseCase) processFile(ctx context.Context, taskEvent dto.TaskEvent) (string, error) {
	fileData, err := uc.objStorage.DownloadFile(ctx, taskEvent.Bucket, taskEvent.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}

	inputFile := dto.InputFile{
		Name: path.Base(taskEvent.FilePath),
		Data: fileData,
	}
	recData, err := uc.recognizer.Recognize(ctx, inputFile)
	if err != nil {
		return "", fmt.Errorf("failed to recognize: %w", err)
	}

	doc := &dto.DocumentObject{
		FileName:   path.Base(taskEvent.FilePath),
		FilePath:   taskEvent.FilePath,
		FileSize:   fileData.Len(),
		Content:    recData.Text,
		CreatedAt:  taskEvent.CreatedAt,
		ModifiedAt: taskEvent.ModifiedAt,
	}

	docID, err := uc.storeToDocSearch(ctx, taskEvent, doc)
	if err != nil {
		return "", fmt.Errorf("failed to store doc to doc-searcher %s: %w", doc.FileName, err)
	}

	slog.Info("successfully stored document",
		slog.String("task-id", taskEvent.ID),
		slog.String("doc-id", docID))
	return docID, nil
}

func (uc *UseCase) storeToDocSearch(ctx context.Context, task dto.TaskEvent, doc *dto.DocumentObject) (string, error) {
	allChunks := uc.textChunker.Chunk(doc.Content)

	rootChunk := allChunks[0]
	splitDoc := &dto.DocumentObject{
		FileName:   doc.FileName,
		FilePath:   doc.FilePath,
		FileSize:   doc.FileSize,
		Content:    rootChunk,
		CreatedAt:  doc.CreatedAt,
		ModifiedAt: doc.ModifiedAt,
	}
	docID, err := uc.docStorage.StoreDocument(ctx, task.Bucket, splitDoc)
	if err != nil {
		return "", fmt.Errorf("failed to store: %w", err)
	}

	chunkSize := len(allChunks)
	slog.Debug("document has been split",
		slog.String("task-id", task.ID),
		slog.String("file-path", doc.FilePath),
		slog.Int("chunks", chunkSize))

	if chunkSize < 2 {
		return docID, nil
	}

	sem := semaphore.NewWeighted(SEMAPHORE_WORKERS_COUNT)
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(allChunks) - 1)
	for _, chunk := range allChunks[1:] {
		chunk := chunk
		go func() {
			defer waitGroup.Done()

			if err := sem.Acquire(ctx, 1); err != nil {
				slog.Error("internal semaphore error",
					slog.String("task-id", task.ID),
					slog.String("err", err.Error()))
				return
			}
			defer sem.Release(1)

			splitDoc := &dto.DocumentObject{
				FileName:   doc.FileName,
				FilePath:   doc.FilePath,
				FileSize:   doc.FileSize,
				Content:    chunk,
				CreatedAt:  doc.CreatedAt,
				ModifiedAt: doc.ModifiedAt,
			}

			docID, err := uc.docStorage.StoreDocument(ctx, task.Bucket, splitDoc)
			if err != nil {
				slog.Error("failed to store document chunk",
					slog.String("task-id", task.ID),
					slog.String("err", err.Error()))
				return
			}

			slog.Info("doc chunk has been stored successful",
				slog.String("task-id", task.ID),
				slog.String("doc-id", docID),
				slog.String("err", err.Error()))
		}()
	}

	waitGroup.Wait()
	return docID, nil
}

func (uc *UseCase) StoreFileToStorage(ctx context.Context, fileForm dto.FileToUpload) (*dto.TaskEvent, error) {
	// TODO: Disabled for TechDebt
	// taskID := utils.GenerateUniqID(fileForm.Bucket, fileForm.FilePath)
	taskID := utils.GenerateTaskID()
	slog.Info("publishing task to queue",
		slog.String("task-taskID", taskID),
		slog.String("index", fileForm.Bucket),
		slog.String("file-path", fileForm.FilePath))

	task := dto.TaskEvent{
		ID:         taskID,
		Bucket:     fileForm.Bucket,
		FilePath:   fileForm.FilePath,
		FileSize:   int64(fileForm.FileData.Len()),
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
		Status:     dto.Received,
	}

	if err := uc.taskManager.Push(ctx, &task); err != nil {
		return nil, fmt.Errorf("failed to store task to cache: %w", err)
	}

	err := uc.objStorage.UploadFile(ctx, fileForm.Bucket, fileForm.FilePath, fileForm.FileData, fileForm.Expired)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	uc.processorCh <- task

	return &task, nil
}

func (uc *UseCase) GetObjectStorage() object_storage.IObjectStorage {
	return uc.objStorage
}

func (uc *UseCase) GetDocStorage() doc_storage.IDocumentStorage {
	return uc.docStorage
}

func (uc *UseCase) GetTaskManager() task_manager.ITaskManager {
	return uc.taskManager
}
