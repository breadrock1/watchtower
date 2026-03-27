package routes_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"watchtower/cmd"
	"watchtower/cmd/watchtower/httpserver/form"
	"watchtower/internal/core/cloud/domain"
	"watchtower/tests/common"
)

const (
	TestBucketName = "test-bucket-name"
	TestBucketPath = "/"

	GetAllBucketsURL = "/api/v1/cloud/buckets"
	CreateBucketURL  = "/api/v1/cloud/bucket"

	GetAllBucketsMethod  = "GetAllBuckets"
	DeleteBucketMethod   = "DeleteBucket"
	CreateBucketMethod   = "CreateBucket"
	IsBucketExistsMethod = "IsBucketExist"
)

var (
	BucketCreatedAt = time.Now()
	TestBucket      = domain.Bucket{
		ID:        TestBucketName,
		Path:      TestBucketPath,
		CreatedAt: BucketCreatedAt,
	}
)

func TestBucketAPIRoutes(t *testing.T) {
	servConfig, err := cmd.InitConfig()
	assert.NoError(t, err, "failed to read config file")

	var getBucketsTestCases = []struct {
		TargetURL           string
		HttpMethod          string
		MockMethodName      string
		ReturnedData        interface{}
		ReturnedError       error
		ExpectedCalledTimes int
		ExpectedStatusCode  int
	}{
		{
			TargetURL:           GetAllBucketsURL,
			HttpMethod:          http.MethodGet,
			MockMethodName:      GetAllBucketsMethod,
			ReturnedData:        []domain.Bucket{TestBucket},
			ReturnedError:       nil,
			ExpectedCalledTimes: 1,
			ExpectedStatusCode:  http.StatusOK,
		},
		{
			TargetURL:           GetAllBucketsURL,
			HttpMethod:          http.MethodGet,
			MockMethodName:      GetAllBucketsMethod,
			ReturnedData:        []domain.Bucket{TestBucket},
			ReturnedError:       fmt.Errorf("failed get all buckets"),
			ExpectedCalledTimes: 1,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
	}

	t.Run("Get buckets", func(t *testing.T) {
		for index, testCase := range getBucketsTestCases {
			testCaseName := fmt.Sprintf("Get buckets case %d", index)
			t.Run(testCaseName, func(t *testing.T) {
				testEnv := common.InitTestAppEnvironment()
				appServer, err := testEnv.BuildAppServer(servConfig)
				assert.NoError(t, err, "failed to build app server")

				testEnv.ObjectStorage.
					On(testCase.MockMethodName).
					Return(testCase.ReturnedData, testCase.ReturnedError)

				req := httptest.NewRequest(testCase.HttpMethod, testCase.TargetURL, nil)

				resp, respErr := appServer.Server.Test(req, -1)
				assert.NoError(t, respErr, "failed to create tag")
				assert.Equal(t, testCase.ExpectedStatusCode, resp.StatusCode, "unexpected http status code")

				testEnv.ObjectStorage.AssertNumberOfCalls(t, testCase.MockMethodName, testCase.ExpectedCalledTimes)
			})
		}
	})

	var createBucketTestCases = []struct {
		TargetURL           string
		HttpMethod          string
		IsBucketExists      bool
		IsBucketExistsError error
		MockMethodName      string
		ReturnedError       error
		RequestPayload      *domain.Bucket
		ExpectedCalledTimes int
		ExpectedStatusCode  int
	}{
		{
			TargetURL:           CreateBucketURL,
			HttpMethod:          http.MethodPut,
			IsBucketExists:      false,
			IsBucketExistsError: nil,
			MockMethodName:      CreateBucketMethod,
			ReturnedError:       nil,
			RequestPayload:      &TestBucket,
			ExpectedCalledTimes: 1,
			ExpectedStatusCode:  http.StatusCreated,
		},
		{
			TargetURL:           CreateBucketURL,
			HttpMethod:          http.MethodPut,
			IsBucketExists:      false,
			IsBucketExistsError: nil,
			MockMethodName:      CreateBucketMethod,
			ReturnedError:       nil,
			RequestPayload:      nil,
			ExpectedCalledTimes: 0,
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			TargetURL:           CreateBucketURL,
			HttpMethod:          http.MethodPut,
			IsBucketExists:      true,
			IsBucketExistsError: nil,
			MockMethodName:      CreateBucketMethod,
			ReturnedError:       nil,
			RequestPayload:      nil,
			ExpectedCalledTimes: 0,
			ExpectedStatusCode:  http.StatusBadRequest,
			// TODO: Temporary implementation
			//ExpectedStatusCode:  http.StatusConflict,
		},
		{
			TargetURL:           CreateBucketURL,
			HttpMethod:          http.MethodPut,
			IsBucketExists:      false,
			IsBucketExistsError: nil,
			MockMethodName:      CreateBucketMethod,
			ReturnedError:       fmt.Errorf("failed create bucket"),
			RequestPayload:      &TestBucket,
			ExpectedCalledTimes: 1,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
	}

	t.Run("Delete bucket", func(t *testing.T) {
		for index, testCase := range createBucketTestCases {
			testCaseName := fmt.Sprintf("Create bucket case %d", index)
			t.Run(testCaseName, func(t *testing.T) {
				testEnv := common.InitTestAppEnvironment()
				appServer, err := testEnv.BuildAppServer(servConfig)
				assert.NoError(t, err, "failed to build app server")

				testEnv.ObjectStorage.
					On(IsBucketExistsMethod, TestBucket.ID).
					Return(testCase.IsBucketExists, testCase.IsBucketExistsError)

				testEnv.ObjectStorage.
					On(testCase.MockMethodName, TestBucket.ID).
					Return(testCase.ReturnedError)

				var buffer = bytes.NewBuffer(nil)
				if testCase.RequestPayload != nil {
					jsonBytes, err := json.Marshal(form.CreateBucketForm{BucketName: testCase.RequestPayload.ID})
					assert.NoError(t, err, "failed to marshal request body")
					buffer = bytes.NewBuffer(jsonBytes)
				}

				req := httptest.NewRequest(testCase.HttpMethod, testCase.TargetURL, buffer)

				resp, respErr := appServer.Server.Test(req, -1)
				assert.NoError(t, respErr, "failed to delete tag")
				assert.Equal(t, testCase.ExpectedStatusCode, resp.StatusCode, "unexpected http status code")

				testEnv.ObjectStorage.AssertNumberOfCalls(t, testCase.MockMethodName, testCase.ExpectedCalledTimes)
			})
		}
	})

	var deleteBucketsTestCases = []struct {
		TargetURL           string
		HttpMethod          string
		MockMethodName      string
		IsBucketExists      bool
		ReturnedError       error
		ExpectedCalledTimes int
		ExpectedStatusCode  int
	}{
		{
			TargetURL:           fmt.Sprintf("/api/v1/cloud/%s", TestBucket.ID),
			HttpMethod:          http.MethodDelete,
			MockMethodName:      DeleteBucketMethod,
			IsBucketExists:      true,
			ReturnedError:       nil,
			ExpectedCalledTimes: 1,
			ExpectedStatusCode:  http.StatusOK,
		},
		{
			TargetURL:           fmt.Sprintf("/api/v1/cloud/%s", TestBucket.ID),
			HttpMethod:          http.MethodDelete,
			MockMethodName:      DeleteBucketMethod,
			IsBucketExists:      false,
			ReturnedError:       nil,
			ExpectedCalledTimes: 0,
			ExpectedStatusCode:  http.StatusNotFound,
		},
		{
			TargetURL:           fmt.Sprintf("/api/v1/cloud/%s", TestBucket.ID),
			HttpMethod:          http.MethodDelete,
			MockMethodName:      DeleteBucketMethod,
			IsBucketExists:      true,
			ReturnedError:       fmt.Errorf("failed to delete bucket"),
			ExpectedCalledTimes: 1,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
	}

	t.Run("Delete buckets", func(t *testing.T) {
		for index, testCase := range deleteBucketsTestCases {
			testCaseName := fmt.Sprintf("Delete bucket case %d", index)
			t.Run(testCaseName, func(t *testing.T) {
				testEnv := common.InitTestAppEnvironment()
				appServer, err := testEnv.BuildAppServer(servConfig)
				assert.NoError(t, err, "failed to build app server")

				testEnv.ObjectStorage.
					On(IsBucketExistsMethod, TestBucket.ID).
					Return(testCase.IsBucketExists, nil)

				testEnv.ObjectStorage.
					On(testCase.MockMethodName, TestBucket.ID).
					Return(testCase.ReturnedError, testCase.ReturnedError)

				req := httptest.NewRequest(testCase.HttpMethod, testCase.TargetURL, nil)

				resp, respErr := appServer.Server.Test(req, -1)
				assert.NoError(t, respErr, "delete bucket")
				assert.Equal(t, testCase.ExpectedStatusCode, resp.StatusCode, "unexpected http status code")

				testEnv.ObjectStorage.AssertNumberOfCalls(t, testCase.MockMethodName, testCase.ExpectedCalledTimes)
			})
		}
	})
}
