package storage_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"watchtower/internal/core/cloud/domain"
	"watchtower/tests/common"
)

const (
	TestBucketName     = "watchtower-test-bucket"
	TestConfigFilePath = "../configs/testing.toml"
)

func TestStorage(t *testing.T) {
	testEnv, initErr := common.InitTestEnvironment(TestConfigFilePath)
	if initErr != nil {
		t.Fatalf("failed to init test environment: %v", initErr)
	}

	t.Run("Generate share URL", func(t *testing.T) {
		ctx := context.Background()

		fileName, err := uuid.NewUUID()
		assert.NoError(t, err, "failed to generate unique file name")

		filePath := fmt.Sprintf("./%s.txt", fileName)
		err = StoreObjectToStorage(ctx, testEnv, filePath)
		assert.NoError(t, err, "failed to upload input file to storage")

		shareParams := domain.ShareObjectParams{
			FilePath: filePath,
			Expired:  10 * time.Minute,
		}

		sharedUrl, err := testEnv.ObjStorage.GenShareURL(ctx, TestBucketName, &shareParams)
		assert.NoError(t, err, "failed to generate shared url from storage")
		assert.Equal(t, sharedUrl.Host, "localhost:9000")
		assert.Equal(t, sharedUrl.Path, fmt.Sprintf("/%s/%s.txt", TestBucketName, fileName))
	})

	t.Run("Generate share URL with prefix", func(t *testing.T) {
		ctx := context.Background()

		fileName, err := uuid.NewUUID()
		assert.NoError(t, err, "failed to generate unique file name")

		filePath := fmt.Sprintf("./any-directory/%s.txt", fileName)
		err = StoreObjectToStorage(ctx, testEnv, filePath)
		assert.NoError(t, err, "failed to upload input file to storage")

		shareParams := domain.ShareObjectParams{
			FilePath: filePath,
			Expired:  10 * time.Minute,
		}

		sharedUrl, err := testEnv.ObjStorage.GenShareURL(ctx, TestBucketName, &shareParams)
		assert.NoError(t, err, "failed to generate shared url from storage")
		assert.Equal(t, sharedUrl.Host, "localhost:9000")
		assert.Equal(t, sharedUrl.Path, fmt.Sprintf("/%s/any-directory/%s.txt", TestBucketName, fileName))
	})
}

func StoreObjectToStorage(ctx context.Context, testEnv *common.TestEnvironment, filePath string) error {
	uploadParams := domain.UploadObjectParams{
		FilePath: filePath,
		FileData: bytes.NewBufferString("there is some file content"),
	}

	_, err := testEnv.ObjStorage.StoreObject(ctx, TestBucketName, &uploadParams)
	return err
}
