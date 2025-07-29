package usecase

import (
	"context"
	"fmt"
	"log"
	"path"
	"time"

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

type UseCase struct {
	processorCh chan dto.TaskEvent
	consumerCh  <-chan dto.Message

	taskQueue   task_queue.ITaskQueue
	taskManager task_manager.ITaskManager
	recognizer  recognizer.IRecognizer
	docStorage  doc_storage.IDocumentStorage
	objStorage  object_storage.IObjectStorage
}

func NewUseCase(
	taskQueue task_queue.ITaskQueue,
	taskManager task_manager.ITaskManager,
	recognizer recognizer.IRecognizer,
	docStorage doc_storage.IDocumentStorage,
	objStorage object_storage.IObjectStorage,
) *UseCase {
	consumerCh := taskQueue.GetConsumerChannel()
	processorCh := make(chan dto.TaskEvent)
	return &UseCase{
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
				if uc.isTaskAlreadyProcessed(ctx, &taskEvent) {
					log.Printf("task has been already processed: %s", taskEvent.ID)
					continue
				}

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

func (uc *UseCase) Processing(ctx context.Context, msg dto.Message) error {
	taskEvent := msg.Body
	uc.updateTaskStatus(ctx, &taskEvent, dto.Processing, EmptyMessage)
	err := uc.processFile(ctx, taskEvent)
	return err
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
		log.Printf("failed to store task to cache: %v", err)
	}
}

func (uc *UseCase) isTaskAlreadyProcessed(ctx context.Context, task *dto.TaskEvent) bool {
	storageTask, err := uc.taskManager.Get(ctx, task.Bucket, task.ID)
	if err != nil {
		log.Printf("failed to get task from cache: %v", err)
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

func (uc *UseCase) processFile(ctx context.Context, taskEvent dto.TaskEvent) error {
	fileData, err := uc.objStorage.DownloadFile(ctx, taskEvent.Bucket, taskEvent.FilePath)
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

	doc := &dto.DocumentObject{
		FileName:   path.Base(taskEvent.FilePath),
		FilePath:   taskEvent.FilePath,
		FileSize:   fileData.Len(),
		Content:    recData.Text,
		CreatedAt:  taskEvent.CreatedAt,
		ModifiedAt: taskEvent.ModifiedAt,
	}

	id, err := uc.docStorage.StoreDocument(ctx, taskEvent.Bucket, doc)
	if err != nil {
		return fmt.Errorf("failed to store doc %s: %w", doc.FileName, err)
	}

	log.Printf("successfully stored document: %s", id)
	return nil
}

func (uc *UseCase) StoreFileToStorage(ctx context.Context, fileForm dto.FileToUpload) (*dto.TaskEvent, error) {
	id := utils.GenerateUniqID(fileForm.Bucket, fileForm.FilePath)
	log.Printf("[%s]: publish task: %s", fileForm.Bucket, id)

	task := dto.TaskEvent{
		ID:         id,
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
