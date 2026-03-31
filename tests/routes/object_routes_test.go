package routes_test

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"watchtower/cmd/watchtower/httpserver/form"
	"watchtower/internal/core/cloud/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"watchtower/cmd"
	"watchtower/tests/common"
)

const (
	TestObjectName        = "test-object-name.docx"
	TestObjectPath        = "./test-object.docx"
	TestObjectNewPath     = "./test/test-object.docx"
	TestObjectContentType = "application/docx"
	TestObjectSize        = 1024
	TestFolderPath        = "test-folder"

	IsBucketExistsMethodName = "IsBucketExist"
	CopyObjectMethodName     = "CopyObject"
	StoreObjectMethodName    = "StoreObject"
	DeleteObjectMethodName   = "DeleteObject"
	DeleteObjectsMethodName  = "DeleteObjects"
)

var (
	TestObjectID        = "test-object-id"
	TestObjectChecksum  = md5.New()
	TestObjectCreatedAt = time.Now()

	TestObject = domain.Object{
		Name:         TestObjectName,
		Path:         TestObjectPath,
		Checksum:     fmt.Sprintf("%x", TestObjectChecksum),
		ContentType:  TestObjectContentType,
		Expired:      TestObjectCreatedAt,
		LastModified: TestObjectCreatedAt,
		Size:         TestObjectSize,
		IsDirectory:  false,
	}

	TestCopyFileForm = form.CopyFileForm{
		SrcPath: TestObjectPath,
		DstPath: TestObjectNewPath,
	}

	TestCreateFolderForm = form.FolderForm{
		Prefix: TestFolderPath,
	}

	MatchedStoreObjectParams = mock.MatchedBy(func(params *domain.UploadObjectParams) bool {
		filePathFlag := params.FilePath == "test-object.docx/.keeper"
		return filePathFlag
	})

	MatchedCopyFilesParams = mock.MatchedBy(func(params *domain.CopyObjectParams) bool {
		srcPathFlag := params.SourcePath == TestObjectPath
		dstPathFlag := params.DestinationPath == TestObjectNewPath
		return srcPathFlag && dstPathFlag
	})
)

// nolint
func TestObjectAPIRoutes(t *testing.T) {
	servConfig, err := cmd.InitConfig()
	assert.NoError(t, err, "failed to read config file")

	var copyFileTestCases = []struct {
		TargetURL           string
		HttpMethod          string
		MockMethodName      string
		RequestPayload      *form.CopyFileForm
		ReturnedData        interface{}
		ReturnedError       error
		ExpectedCalledTimes int
		ExpectedStatusCode  int
	}{
		{
			TargetURL:           fmt.Sprintf("/api/v1/cloud/%s/file/copy", TestBucketName),
			HttpMethod:          http.MethodPost,
			MockMethodName:      CopyObjectMethodName,
			RequestPayload:      &TestCopyFileForm,
			ReturnedData:        nil,
			ReturnedError:       nil,
			ExpectedCalledTimes: 1,
			ExpectedStatusCode:  http.StatusOK,
		},
		{
			TargetURL:           fmt.Sprintf("/api/v1/cloud/%s/file/copy", TestBucketName),
			HttpMethod:          http.MethodPost,
			MockMethodName:      CopyObjectMethodName,
			RequestPayload:      nil,
			ReturnedData:        nil,
			ReturnedError:       errors.New("invalid request"),
			ExpectedCalledTimes: 0,
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			TargetURL:           fmt.Sprintf("/api/v1/cloud/%s/file/copy", TestBucketName),
			HttpMethod:          http.MethodPost,
			MockMethodName:      CopyObjectMethodName,
			RequestPayload:      &TestCopyFileForm,
			ReturnedData:        nil,
			ReturnedError:       errors.New("internal error"),
			ExpectedCalledTimes: 1,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
	}

	t.Run("Copy file", func(t *testing.T) {
		for index, testCase := range copyFileTestCases {
			testCaseName := fmt.Sprintf("Copy file case %d", index)
			t.Run(testCaseName, func(t *testing.T) {
				testEnv := common.InitTestAppEnvironment()
				appServer, err := testEnv.BuildAppServer(servConfig)
				assert.NoError(t, err, "failed to build app server")

				testEnv.ObjectStorage.
					On(testCase.MockMethodName, TestBucketName, MatchedCopyFilesParams).
					Return(testCase.ReturnedError)

				var buffer = bytes.NewBuffer(nil)
				if testCase.RequestPayload != nil {
					jsonBytes, err := json.Marshal(testCase.RequestPayload)
					assert.NoError(t, err, "failed to marshal request body")
					buffer = bytes.NewBuffer(jsonBytes)
				}

				req := httptest.NewRequest(testCase.HttpMethod, testCase.TargetURL, buffer)

				resp, respErr := appServer.Server.Test(req, -1)
				assert.NoError(t, respErr, "failed to copy file")
				assert.Equal(t, testCase.ExpectedStatusCode, resp.StatusCode, "unexpected http status code")

				testEnv.ObjectStorage.AssertNumberOfCalls(t, testCase.MockMethodName, testCase.ExpectedCalledTimes)
			})
		}
	})

	var createFolderTestCases = []struct {
		TargetURL           string
		HttpMethod          string
		RequestPayload      *form.FolderForm
		IsBucketExists      bool
		MockMethodName      string
		ReturnedData        interface{}
		ReturnedError       error
		ExpectedCalledTimes int
		ExpectedStatusCode  int
	}{
		{
			TargetURL:           fmt.Sprintf("/api/v1/cloud/%s/folder", TestBucketName),
			HttpMethod:          http.MethodPost,
			RequestPayload:      &TestCreateFolderForm,
			IsBucketExists:      true,
			MockMethodName:      StoreObjectMethodName,
			ReturnedData:        nil,
			ReturnedError:       nil,
			ExpectedCalledTimes: 1,
			ExpectedStatusCode:  http.StatusCreated,
		},
		{
			TargetURL:           fmt.Sprintf("/api/v1/cloud/%s/folder", TestBucketName),
			HttpMethod:          http.MethodPost,
			RequestPayload:      &TestCreateFolderForm,
			IsBucketExists:      false,
			MockMethodName:      StoreObjectMethodName,
			ReturnedData:        nil,
			ReturnedError:       nil,
			ExpectedCalledTimes: 0,
			ExpectedStatusCode:  http.StatusNotFound,
		},
		{
			TargetURL:           fmt.Sprintf("/api/v1/cloud/%s/folder", TestBucketName),
			HttpMethod:          http.MethodPost,
			RequestPayload:      nil,
			IsBucketExists:      true,
			MockMethodName:      StoreObjectMethodName,
			ReturnedData:        nil,
			ReturnedError:       errors.New("invalid folder name"),
			ExpectedCalledTimes: 0,
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			TargetURL:           fmt.Sprintf("/api/v1/cloud/%s/folder", TestBucketName),
			HttpMethod:          http.MethodPost,
			RequestPayload:      &TestCreateFolderForm,
			IsBucketExists:      true,
			MockMethodName:      StoreObjectMethodName,
			ReturnedData:        nil,
			ReturnedError:       errors.New("internal server error"),
			ExpectedCalledTimes: 1,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
	}

	//nolint
	t.Run("Create folder", func(t *testing.T) {
		for index, testCase := range createFolderTestCases {
			testCaseName := fmt.Sprintf("Create folder case %d", index)
			t.Run(testCaseName, func(t *testing.T) {
				testEnv := common.InitTestAppEnvironment()
				appServer, err := testEnv.BuildAppServer(servConfig)
				assert.NoError(t, err, "failed to build app server")

				testEnv.ObjectStorage.
					On(IsBucketExistsMethodName, TestBucketName).
					Return(testCase.IsBucketExists, nil)

				testEnv.ObjectStorage.
					On(testCase.MockMethodName, TestBucketName, MatchedStoreObjectParams).
					Return(TestObjectID, testCase.ReturnedError)

				var buffer = bytes.NewBuffer(nil)
				if testCase.RequestPayload != nil {
					jsonBytes, err := json.Marshal(testCase.RequestPayload)
					assert.NoError(t, err, "failed to marshal request body")
					buffer = bytes.NewBuffer(jsonBytes)
				}

				req := httptest.NewRequest(testCase.HttpMethod, testCase.TargetURL, buffer)

				resp, respErr := appServer.Server.Test(req, -1)
				assert.NoError(t, respErr, "failed to create folder")
				assert.Equal(t, testCase.ExpectedStatusCode, resp.StatusCode, "unexpected http status code")

				testEnv.ObjectStorage.AssertNumberOfCalls(t, testCase.MockMethodName, testCase.ExpectedCalledTimes)
			})
		}
	})

	var deleteFolderTestCases = []struct {
		TargetURL           string
		HttpMethod          string
		RequestPayload      *form.FolderForm
		IsBucketExists      bool
		MockMethodName      string
		ReturnedData        interface{}
		ReturnedError       error
		ExpectedCalledTimes int
		ExpectedStatusCode  int
	}{
		{
			TargetURL:           fmt.Sprintf("/api/v1/cloud/%s/folder", TestBucketName),
			HttpMethod:          http.MethodDelete,
			RequestPayload:      &TestCreateFolderForm,
			IsBucketExists:      true,
			MockMethodName:      DeleteObjectsMethodName,
			ReturnedData:        nil,
			ReturnedError:       nil,
			ExpectedCalledTimes: 1,
			ExpectedStatusCode:  http.StatusOK,
		},
		{
			TargetURL:           fmt.Sprintf("/api/v1/cloud/%s/folder", TestBucketName),
			HttpMethod:          http.MethodDelete,
			RequestPayload:      &TestCreateFolderForm,
			IsBucketExists:      false,
			MockMethodName:      DeleteObjectsMethodName,
			ReturnedData:        nil,
			ReturnedError:       errors.New("bucket not found"),
			ExpectedCalledTimes: 0,
			ExpectedStatusCode:  http.StatusNotFound,
		},
		{
			TargetURL:           fmt.Sprintf("/api/v1/cloud/%s/folder", TestBucketName),
			HttpMethod:          http.MethodDelete,
			RequestPayload:      &TestCreateFolderForm,
			IsBucketExists:      true,
			MockMethodName:      DeleteObjectMethodName,
			ReturnedData:        nil,
			ReturnedError:       errors.New("folder not empty"),
			ExpectedCalledTimes: 0,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
		{
			TargetURL:           fmt.Sprintf("/api/v1/cloud/%s/folder", TestBucketName),
			HttpMethod:          http.MethodDelete,
			RequestPayload:      &TestCreateFolderForm,
			IsBucketExists:      true,
			MockMethodName:      DeleteObjectMethodName,
			ReturnedData:        nil,
			ReturnedError:       errors.New("service unavailable"),
			ExpectedCalledTimes: 0,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
	}

	//nolint
	t.Run("Delete folder", func(t *testing.T) {
		for index, testCase := range deleteFolderTestCases {
			testCaseName := fmt.Sprintf("Delete folder case %d", index)
			t.Run(testCaseName, func(t *testing.T) {
				testEnv := common.InitTestAppEnvironment()
				appServer, err := testEnv.BuildAppServer(servConfig)
				assert.NoError(t, err, "failed to build app server")

				testEnv.ObjectStorage.
					On(IsBucketExistsMethodName, TestBucketName).
					Return(testCase.IsBucketExists, nil)

				testEnv.ObjectStorage.
					On(testCase.MockMethodName, TestBucketName, TestFolderPath).
					Return(TestObjectID, testCase.ReturnedError)

				var buffer = bytes.NewBuffer(nil)
				if testCase.RequestPayload != nil {
					jsonBytes, err := json.Marshal(testCase.RequestPayload)
					assert.NoError(t, err, "failed to marshal request body")
					buffer = bytes.NewBuffer(jsonBytes)
				}

				req := httptest.NewRequest(testCase.HttpMethod, testCase.TargetURL, buffer)

				resp, respErr := appServer.Server.Test(req, -1)
				assert.NoError(t, respErr, "failed to delete folder")
				assert.Equal(t, testCase.ExpectedStatusCode, resp.StatusCode, "unexpected http status code")

				testEnv.ObjectStorage.AssertNumberOfCalls(t, testCase.MockMethodName, testCase.ExpectedCalledTimes)
			})
		}
	})
}
