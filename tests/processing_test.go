package processing_test

import (
	"context"
	"fmt"
	"path"
	"testing"
	"time"
	domain "watchtower/internal/domain/core/structures"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"watchtower/internal/application/mapping"
	"watchtower/internal/application/models"
	"watchtower/tests/common"
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
		testEnv.PipelineUC.LaunchListener(cCtx)

		fileForm, err := common.CreateUploadFileParams(TestInputFilePath)
		if err != nil {
			t.Fatalf("failed to create upload params: %v", err)
		}

		docID := uuid.New().String()
		recData := models.Recognized{Text: fileForm.FileData.String()}
		inputFile := models.InputFile{Name: fileForm.FilePath, Data: *fileForm.FileData}
		docObject := models.DocumentObject{
			FileName:   path.Base(fileForm.FilePath),
			FilePath:   fileForm.FilePath,
			FileSize:   fileForm.FileData.Len(),
			Content:    fileForm.FileData.String(),
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
		}
		matchedStoreDocument := mock.MatchedBy(func(doc *models.DocumentObject) bool {
			fileNameFlag := doc.FileName == docObject.FileName
			filePathFlag := doc.FilePath == docObject.FilePath
			fileSizeFlag := doc.FileSize == docObject.FileSize
			contentFlag := doc.Content == docObject.Content
			return fileNameFlag && filePathFlag && fileSizeFlag && contentFlag
		})
		testEnv.Recognizer.On("Recognize", inputFile).Return(&recData, nil).Once()
		testEnv.DocStorage.On("StoreDocument", TestBucketName, matchedStoreDocument).Return(docID, nil).Once()

		task, err := testEnv.PipelineUC.CreateTask(ctx, *fileForm)
		assert.NoError(t, err, "failed to upload test input file to s3")

		timeoutCh := time.After(7 * time.Second)
		<-timeoutCh

		testEnv.Recognizer.AssertExpectations(t)
		testEnv.DocStorage.AssertExpectations(t)

		task, err = testEnv.TaskManager.Get(ctx, TestBucketName, task.ID)
		assert.NoError(t, err, "failed to get task from redis")
		assert.Equal(t, mapping.TaskStatusFromInt(3), task.Status)
		assert.Equal(t, TestBucketName, task.Bucket)
		assert.Equal(t, TestInputFilePath, task.FilePath)

		cancel()
	})

	t.Run("Failed to load object pipeline", func(t *testing.T) {
		ctx := context.Background()
		cCtx, cancel := context.WithCancel(ctx)
		testEnv.PipelineUC.LaunchListener(cCtx)

		taskID := domain.GenerateUniqID(TestBucketName, TestInputFilePath)
		taskEvent := models.TaskEvent{
			ID:         taskID,
			Bucket:     TestBucketName,
			FilePath:   path.Base(TestInputFilePath),
			FileSize:   0,
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
			Status:     1,
			StatusText: "",
		}

		rmqMsg := mapping.MessageFromTaskEvent(taskEvent)
		err := testEnv.TaskQueue.Publish(ctx, rmqMsg)
		assert.NoError(t, err, "failed to publish task event")

		timeoutCh := time.After(7 * time.Second)
		<-timeoutCh

		testEnv.Recognizer.AssertNotCalled(t, "Recognize")
		testEnv.DocStorage.AssertNotCalled(t, "StoreDocument")

		task, err := testEnv.TaskManager.Get(ctx, TestBucketName, taskID)
		assert.NoError(t, err, "failed to get task from redis")
		assert.Equal(t, mapping.TaskStatusFromInt(-1), task.Status)
		assert.Equal(t, TestBucketName, task.Bucket)
		assert.Equal(t, path.Base(TestInputFilePath), task.FilePath)

		cancel()
	})

	t.Run("Failed to recognize pipeline", func(t *testing.T) {
		ctx := context.Background()
		cCtx, cancel := context.WithCancel(ctx)
		testEnv.PipelineUC.LaunchListener(cCtx)

		fileForm, err := common.CreateUploadFileParams(TestInputFilePath)
		if err != nil {
			t.Fatalf("failed to create upload params: %v", err)
		}

		recData := models.Recognized{Text: fileForm.FileData.String()}
		inputFile := models.InputFile{Name: fileForm.FilePath, Data: *fileForm.FileData}
		recErr := fmt.Errorf("service unavailable")
		testEnv.Recognizer.On("Recognize", inputFile).Return(&recData, recErr).Once()

		task, err := testEnv.PipelineUC.CreateTask(ctx, *fileForm)
		assert.NoError(t, err, "failed to upload test input file to s3")

		timeoutCh := time.After(7 * time.Second)
		<-timeoutCh

		testEnv.Recognizer.AssertExpectations(t)
		testEnv.DocStorage.AssertNotCalled(t, "StoreDocument")

		task, err = testEnv.TaskManager.Get(ctx, TestBucketName, task.ID)
		assert.NoError(t, err, "failed to get task from redis")
		assert.Equal(t, mapping.TaskStatusFromInt(-1), task.Status)
		assert.Equal(t, TestBucketName, task.Bucket)
		assert.Equal(t, TestInputFilePath, task.FilePath)

		cancel()
	})

	t.Run("Failed to store document pipeline", func(t *testing.T) {
		ctx := context.Background()
		cCtx, cancel := context.WithCancel(ctx)
		testEnv.PipelineUC.LaunchListener(cCtx)

		fileForm, err := common.CreateUploadFileParams(TestInputFilePath)
		if err != nil {
			t.Fatalf("failed to create upload params: %v", err)
		}

		recData := models.Recognized{Text: fileForm.FileData.String()}
		inputFile := models.InputFile{Name: fileForm.FilePath, Data: *fileForm.FileData}
		docObject := models.DocumentObject{
			FileName:   path.Base(fileForm.FilePath),
			FilePath:   fileForm.FilePath,
			FileSize:   fileForm.FileData.Len(),
			Content:    fileForm.FileData.String(),
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
		}
		matchedStoreDocument := mock.MatchedBy(func(doc *models.DocumentObject) bool {
			fileNameFlag := doc.FileName == docObject.FileName
			filePathFlag := doc.FilePath == docObject.FilePath
			fileSizeFlag := doc.FileSize == docObject.FileSize
			contentFlag := doc.Content == docObject.Content
			return fileNameFlag && filePathFlag && fileSizeFlag && contentFlag
		})
		docErr := fmt.Errorf("service unavailable")
		testEnv.Recognizer.On("Recognize", inputFile).Return(&recData, nil).Once()
		testEnv.DocStorage.On("StoreDocument", TestBucketName, matchedStoreDocument).Return("", docErr).Once()

		task, err := testEnv.PipelineUC.CreateTask(ctx, *fileForm)
		assert.NoError(t, err, "failed to upload test input file to s3")

		timeoutCh := time.After(7 * time.Second)
		<-timeoutCh

		testEnv.Recognizer.AssertExpectations(t)
		testEnv.DocStorage.AssertExpectations(t)

		task, err = testEnv.TaskManager.Get(ctx, TestBucketName, task.ID)
		assert.NoError(t, err, "failed to get task from redis")
		assert.Equal(t, mapping.TaskStatusFromInt(-1), task.Status)
		assert.Equal(t, TestBucketName, task.Bucket)
		assert.Equal(t, TestInputFilePath, task.FilePath)

		cancel()
	})
}
