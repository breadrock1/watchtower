package processing_test

import (
	"context"
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"watchtower/internal/support/task/application/mapping"
	"watchtower/internal/support/task/application/service/docstorage"
	"watchtower/internal/support/task/application/service/recognizer"
	"watchtower/tests/common"

	taskDomain "watchtower/internal/support/task/domain"
)

const (
	TestBucketName     = "watchtower-test-bucket"
	TestInputFilePath  = "./resources/input-file.txt"
	TestConfigFilePath = "../configs/testing.toml"
)

func TestProcessing(t *testing.T) {
	testEnv, initErr := common.InitTestEnvironment(TestConfigFilePath)
	if initErr != nil {
		t.Fatalf("failed to init test environment: %v", initErr)
	}

	t.Run("Positive pipeline", func(t *testing.T) {
		ctx := context.Background()
		cCtx, cancel := context.WithCancel(ctx)
		testEnv.Orchestrator.LaunchListener(cCtx)

		fileForm, err := common.CreateUploadFileParams(TestInputFilePath)
		if err != nil {
			t.Fatalf("failed to create upload params: %v", err)
		}

		docID := uuid.New().String()
		recData := recognizer.Recognized{Text: fileForm.FileData.String()}
		recParams := recognizer.RecognizeParams{FileName: fileForm.FilePath, FileData: fileForm.FileData}
		docObject := docstorage.Document{
			Index:      TestBucketName,
			Name:       path.Base(fileForm.FilePath),
			Path:       fileForm.FilePath,
			Size:       fileForm.FileData.Len(),
			Content:    fileForm.FileData.String(),
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
		}
		matchedStoreDocument := mock.MatchedBy(func(doc docstorage.Document) bool {
			indexFlag := doc.Index == TestBucketName
			fileNameFlag := doc.Name == docObject.Name
			filePathFlag := doc.Path == docObject.Path
			fileSizeFlag := doc.Size == docObject.Size
			contentFlag := doc.Content == docObject.Content
			return indexFlag && fileNameFlag && filePathFlag && fileSizeFlag && contentFlag
		})
		testEnv.Recognizer.On("Recognize", recParams).Return(recData, nil).Once()
		testEnv.DocStorage.On("StoreDocument", matchedStoreDocument).Return(docID, nil).Once()

		task, err := testEnv.Orchestrator.CreateTask(ctx, docObject.Index, fileForm.FilePath)
		assert.NoError(t, err, "failed to upload test input file to s3")

		timeoutCh := time.After(7 * time.Second)
		<-timeoutCh

		testEnv.Recognizer.AssertExpectations(t)
		testEnv.DocStorage.AssertExpectations(t)

		loadTask, err := testEnv.TaskManager.GetTask(ctx, task.BucketID, task.ID)
		assert.NoError(t, err, "failed to get task from redis")
		assert.Equal(t, taskDomain.Successful, loadTask.Status)
		assert.Equal(t, task.BucketID, loadTask.BucketID)
		assert.Equal(t, task.ObjectID, loadTask.ObjectID)

		cancel()
	})

	t.Run("Failed to load object pipeline", func(t *testing.T) {
		ctx := context.Background()
		cCtx, cancel := context.WithCancel(ctx)
		testEnv.Orchestrator.LaunchListener(cCtx)

		task := taskDomain.CreateNewTask(TestBucketName, path.Base(TestInputFilePath))

		rmqMsg := mapping.MessageFromTask(task)
		err := testEnv.TaskQueue.Publish(ctx, rmqMsg)
		assert.NoError(t, err, "failed to publish task event")

		timeoutCh := time.After(7 * time.Second)
		<-timeoutCh

		testEnv.Recognizer.AssertNotCalled(t, "Recognize")
		testEnv.DocStorage.AssertNotCalled(t, "StoreDocument")

		loadTask, err := testEnv.TaskManager.GetTask(ctx, task.BucketID, task.ID)
		assert.NoError(t, err, "failed to get task from redis")
		assert.Equal(t, taskDomain.Failed, loadTask.Status)
		assert.Equal(t, task.BucketID, loadTask.BucketID)
		assert.Equal(t, task.ObjectID, loadTask.ObjectID)

		cancel()
	})

	t.Run("Failed to recognize pipeline", func(t *testing.T) {
		ctx := context.Background()
		cCtx, cancel := context.WithCancel(ctx)
		testEnv.Orchestrator.LaunchListener(cCtx)

		fileForm, err := common.CreateUploadFileParams(TestInputFilePath)
		if err != nil {
			t.Fatalf("failed to create upload params: %v", err)
		}

		recData := recognizer.Recognized{Text: fileForm.FileData.String()}
		recParams := recognizer.RecognizeParams{FileName: fileForm.FilePath, FileData: fileForm.FileData}
		recErr := fmt.Errorf("service unavailable")
		testEnv.Recognizer.On("Recognize", recParams).Return(recData, recErr).Once()

		task, err := testEnv.Orchestrator.CreateTask(ctx, TestBucketName, fileForm.FilePath)
		assert.NoError(t, err, "failed to upload test input file to s3")

		timeoutCh := time.After(7 * time.Second)
		<-timeoutCh

		testEnv.Recognizer.AssertExpectations(t)
		testEnv.DocStorage.AssertNotCalled(t, "StoreDocument")

		loadTask, err := testEnv.TaskManager.GetTask(ctx, task.BucketID, task.ID)
		assert.NoError(t, err, "failed to get task from redis")
		assert.Equal(t, taskDomain.Failed, loadTask.Status)
		assert.Equal(t, task.BucketID, loadTask.BucketID)
		assert.Equal(t, task.ObjectID, loadTask.ObjectID)

		cancel()
	})

	t.Run("Failed to store document pipeline", func(t *testing.T) {
		ctx := context.Background()
		cCtx, cancel := context.WithCancel(ctx)
		testEnv.Orchestrator.LaunchListener(cCtx)

		fileForm, err := common.CreateUploadFileParams(TestInputFilePath)
		if err != nil {
			t.Fatalf("failed to create upload params: %v", err)
		}

		recData := recognizer.Recognized{Text: fileForm.FileData.String()}
		recParams := recognizer.RecognizeParams{FileName: fileForm.FilePath, FileData: fileForm.FileData}
		docObject := docstorage.Document{
			Index:      TestBucketName,
			Name:       path.Base(fileForm.FilePath),
			Path:       fileForm.FilePath,
			Size:       fileForm.FileData.Len(),
			Content:    fileForm.FileData.String(),
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
		}
		matchedStoreDocument := mock.MatchedBy(func(doc docstorage.Document) bool {
			indexFlag := doc.Index == TestBucketName
			fileNameFlag := doc.Name == docObject.Name
			filePathFlag := doc.Path == docObject.Path
			fileSizeFlag := doc.Size == docObject.Size
			contentFlag := doc.Content == docObject.Content
			return indexFlag && fileNameFlag && filePathFlag && fileSizeFlag && contentFlag
		})
		docErr := fmt.Errorf("service unavailable")
		testEnv.Recognizer.On("Recognize", recParams).Return(recData, nil).Once()
		testEnv.DocStorage.On("StoreDocument", matchedStoreDocument).Return("", docErr).Once()

		task, err := testEnv.Orchestrator.CreateTask(ctx, TestBucketName, fileForm.FilePath)
		assert.NoError(t, err, "failed to upload test input file to s3")

		timeoutCh := time.After(7 * time.Second)
		<-timeoutCh

		testEnv.Recognizer.AssertExpectations(t)
		testEnv.DocStorage.AssertExpectations(t)

		loadTask, err := testEnv.TaskManager.GetTask(ctx, task.BucketID, task.ID)
		assert.NoError(t, err, "failed to get task from redis")
		assert.Equal(t, taskDomain.Failed, loadTask.Status)
		assert.Equal(t, task.BucketID, loadTask.BucketID)
		assert.Equal(t, task.ObjectID, loadTask.ObjectID)

		cancel()
	})
}
