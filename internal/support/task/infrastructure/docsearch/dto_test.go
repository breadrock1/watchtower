package docsearch

import (
	"fmt"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testFileName       = "file.txt"
	testEmptyFileName  = "empty.txt"
	testFilePath       = "dir/file.txt"
	testFileSize       = 4
	testContent        = "text"
	testUnicodeContent = "First line\nSecond line"
	testCreatedAt      = int64(10)
	testModifiedAt     = int64(20)
	testSuccessStatus  = 200
	testErrorStatus    = 500
	testDocumentID     = "doc-id"

	testDocumentMessageJSON    = `{"message":"doc-id"}`
	testUnknownFieldResultJSON = `{"status":200,"message":"doc-id","extra":"ignored"}`
)

func TestStoreDocumentDTO(t *testing.T) {
	t.Run("Store request and response fields", func(t *testing.T) {
		form := StoreDocumentForm{
			FileName:   testFileName,
			FilePath:   testFilePath,
			FileSize:   testFileSize,
			Content:    testContent,
			CreatedAt:  testCreatedAt,
			ModifiedAt: testModifiedAt,
		}
		result := StoreDocumentResult{Status: testSuccessStatus, Message: testDocumentID}

		assert.Equal(t, testFileName, form.FileName, "unexpected file name")
		assert.Equal(t, testFilePath, form.FilePath, "unexpected file path")
		assert.Equal(t, testFileSize, form.FileSize, "unexpected file size")
		assert.Equal(t, testContent, form.Content, "unexpected content")
		assert.Equal(t, testCreatedAt, form.CreatedAt, "unexpected creation timestamp")
		assert.Equal(t, testModifiedAt, form.ModifiedAt, "unexpected modification timestamp")
		assert.Equal(t, testSuccessStatus, result.Status, "unexpected result status")
		assert.Equal(t, testDocumentID, result.Message, "unexpected result message")
	})

	t.Run("Marshal store document form to JSON", func(t *testing.T) {
		form := StoreDocumentForm{
			FileName:   testFileName,
			FilePath:   testFilePath,
			FileSize:   testFileSize,
			Content:    testContent,
			CreatedAt:  testCreatedAt,
			ModifiedAt: testModifiedAt,
		}

		data, err := json.Marshal(form)
		require.NoError(t, err, "expected form to marshal")

		var got map[string]any
		err = json.Unmarshal(data, &got)
		require.NoError(t, err, "expected marshaled form to unmarshal")

		assert.Equal(t, testFileName, got["file_name"], "unexpected JSON file name")
		assert.Equal(t, testFilePath, got["file_path"], "unexpected JSON file path")
		assert.Equal(t, float64(testFileSize), got["file_size"], "unexpected JSON file size")
		assert.Equal(t, testContent, got["content"], "unexpected JSON content")
		assert.Equal(t, float64(testCreatedAt), got["created_at"], "unexpected JSON created timestamp")
		assert.Equal(t, float64(testModifiedAt), got["modified_at"], "unexpected JSON modified timestamp")
	})
	t.Run("Unmarshal success status", func(t *testing.T) {
		data := []byte(fmt.Sprintf(`{"status":%d}`, testSuccessStatus))

		var result StoreDocumentResult
		err := json.Unmarshal(data, &result)

		require.NoError(t, err, "expected result to unmarshal")
		assert.Equal(t, testSuccessStatus, result.Status, "unexpected result status")
	})
	t.Run("Unmarshal error status", func(t *testing.T) {
		data := []byte(fmt.Sprintf(`{"status":%d}`, testErrorStatus))

		var result StoreDocumentResult
		err := json.Unmarshal(data, &result)

		require.NoError(t, err, "expected result to unmarshal")
		assert.Equal(t, testErrorStatus, result.Status, "unexpected result status")
	})

	t.Run("Marshal empty store document form", func(t *testing.T) {
		form := StoreDocumentForm{}

		data, err := json.Marshal(form)
		require.NoError(t, err, "expected empty form to marshal")

		var got map[string]any
		err = json.Unmarshal(data, &got)
		require.NoError(t, err, "expected empty form JSON to unmarshal")

		assert.Equal(t, "", got["file_name"], "unexpected JSON file name")
		assert.Equal(t, "", got["file_path"], "unexpected JSON file path")
		assert.Equal(t, float64(0), got["file_size"], "unexpected JSON file size")
		assert.Equal(t, "", got["content"], "unexpected JSON content")
		assert.Equal(t, float64(0), got["created_at"], "unexpected JSON created timestamp")
		assert.Equal(t, float64(0), got["modified_at"], "unexpected JSON modified timestamp")
	})

	t.Run("Marshal content with unicode and newlines", func(t *testing.T) {
		form := StoreDocumentForm{
			FileName: testFileName,
			Content:  testUnicodeContent,
		}

		data, err := json.Marshal(form)
		require.NoError(t, err, "expected form to marshal")

		var got StoreDocumentForm
		err = json.Unmarshal(data, &got)
		require.NoError(t, err, "expected form to unmarshal")

		assert.Equal(t, form.Content, got.Content, "expected content to be preserved")
	})

	t.Run("Marshal zero file size", func(t *testing.T) {
		form := StoreDocumentForm{
			FileName: testEmptyFileName,
			FileSize: 0,
		}

		data, err := json.Marshal(form)
		require.NoError(t, err, "expected form to marshal")

		var got StoreDocumentForm
		err = json.Unmarshal(data, &got)
		require.NoError(t, err, "expected form to unmarshal")

		assert.Equal(t, 0, got.FileSize, "unexpected file size")
	})

	t.Run("Unmarshal partial store document result", func(t *testing.T) {
		data := []byte(testDocumentMessageJSON)

		var result StoreDocumentResult
		err := json.Unmarshal(data, &result)
		require.NoError(t, err, "expected partial result to unmarshal")

		assert.Equal(t, 0, result.Status, "expected zero status")
		assert.Equal(t, testDocumentID, result.Message, "unexpected result message")
	})

	t.Run("Ignore unknown result fields", func(t *testing.T) {
		data := []byte(testUnknownFieldResultJSON)

		var result StoreDocumentResult
		err := json.Unmarshal(data, &result)
		require.NoError(t, err, "expected result with extra field to unmarshal")

		assert.Equal(t, testSuccessStatus, result.Status, "unexpected result status")
		assert.Equal(t, testDocumentID, result.Message, "unexpected result message")
	})
}
