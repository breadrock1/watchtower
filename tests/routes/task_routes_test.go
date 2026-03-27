package routes_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"watchtower/cmd"
	"watchtower/internal/shared/kernel"
	"watchtower/internal/support/task/domain"
	"watchtower/tests/common"
)

const (
	TestTaskStatus     = "done"
	TestObjectDataSize = 1024

	IncorrectTaskID = "incorrect-task-id"

	GetTaskMethod   = "GetTask"
	LoadTasksMethod = "GetAllBucketTasks"
)

var (
	TestTaskID      = uuid.New()
	TestTaskCreated = time.Now()
	TestTask        = domain.Task{
		ID:                 TestTaskID,
		BucketID:           TestBucket.ID,
		ObjectID:           TestObjectID,
		ObjectDataSize:     TestObjectDataSize,
		StatusText:         TestTaskStatus,
		Status:             domain.Successful,
		CreatedAt:          TestTaskCreated,
		ModifiedAt:         TestTaskCreated,
		RetryCount:         0,
		MaxRetries:         0,
		ProcessingDuration: 1 * time.Second,
	}

	matchedBucketID = mock.MatchedBy(func(id kernel.BucketID) bool {
		return id == TestBucket.ID
	})

	matchedTaskID = mock.MatchedBy(func(id kernel.TaskID) bool {
		return id.String() == TestTaskID.String()
	})
)

func TestTaskAPIRoutes(t *testing.T) {
	servConfig, err := cmd.InitConfig()
	assert.NoError(t, err, "failed to read config file")

	var loadTasksTestCases = []struct {
		TargetURL           string
		HttpMethod          string
		MockMethodName      string
		ReturnedData        interface{}
		ReturnedError       error
		ExpectedCalledTimes int
		ExpectedStatusCode  int
	}{
		{
			TargetURL:           fmt.Sprintf("/api/v1/tasks/%s?status=0", TestBucketName),
			HttpMethod:          http.MethodGet,
			MockMethodName:      LoadTasksMethod,
			ReturnedData:        []*domain.Task{&TestTask},
			ReturnedError:       nil,
			ExpectedCalledTimes: 1,
			ExpectedStatusCode:  http.StatusOK,
		},
		{
			TargetURL:           fmt.Sprintf("/api/v1/tasks/%s", TestBucketName),
			HttpMethod:          http.MethodGet,
			MockMethodName:      LoadTasksMethod,
			ReturnedData:        nil,
			ReturnedError:       nil,
			ExpectedCalledTimes: 0,
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			TargetURL:           fmt.Sprintf("/api/v1/tasks/%s?status=kek", TestBucketName),
			HttpMethod:          http.MethodGet,
			MockMethodName:      LoadTasksMethod,
			ReturnedData:        nil,
			ReturnedError:       nil,
			ExpectedCalledTimes: 0,
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			TargetURL:           fmt.Sprintf("/api/v1/tasks/%s?status=0", TestBucketName),
			HttpMethod:          http.MethodGet,
			MockMethodName:      LoadTasksMethod,
			ReturnedData:        []*domain.Task{&TestTask},
			ReturnedError:       fmt.Errorf("failed to load tasks"),
			ExpectedCalledTimes: 1,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
	}

	t.Run("Load tasks", func(t *testing.T) {
		for index, testCase := range loadTasksTestCases {
			testCaseName := fmt.Sprintf("Load tasks case %d", index)
			t.Run(testCaseName, func(t *testing.T) {
				testEnv := common.InitTestAppEnvironment()
				appServer, err := testEnv.BuildAppServer(servConfig)
				assert.NoError(t, err, "failed to build app server")

				testEnv.TaskStorage.
					On(testCase.MockMethodName, matchedBucketID).
					Return(testCase.ReturnedData, testCase.ReturnedError)

				req := httptest.NewRequest(testCase.HttpMethod, testCase.TargetURL, nil)

				resp, respErr := appServer.Server.Test(req, -1)
				assert.NoError(t, respErr, "failed to load tasks")
				assert.Equal(t, testCase.ExpectedStatusCode, resp.StatusCode, "unexpected http status code")

				testEnv.TaskStorage.AssertNumberOfCalls(t, testCase.MockMethodName, testCase.ExpectedCalledTimes)
			})
		}
	})

	var loadTaskByIDTestCases = []struct {
		TargetURL           string
		HttpMethod          string
		MockMethodName      string
		ReturnedData        interface{}
		ReturnedError       error
		ExpectedCalledTimes int
		ExpectedStatusCode  int
	}{
		{
			TargetURL:           fmt.Sprintf("/api/v1/tasks/%s/%s", TestBucketName, TestTaskID.String()),
			HttpMethod:          http.MethodGet,
			MockMethodName:      GetTaskMethod,
			ReturnedData:        &TestTask,
			ReturnedError:       nil,
			ExpectedCalledTimes: 1,
			ExpectedStatusCode:  http.StatusOK,
		},
		{
			TargetURL:           fmt.Sprintf("/api/v1/tasks/%s/%s", TestBucketName, IncorrectTaskID),
			HttpMethod:          http.MethodGet,
			MockMethodName:      GetTaskMethod,
			ReturnedData:        nil,
			ReturnedError:       nil,
			ExpectedCalledTimes: 0,
			ExpectedStatusCode:  http.StatusBadRequest,
		},
		{
			TargetURL:           fmt.Sprintf("/api/v1/tasks/%s/%s", TestBucketName, TestTaskID.String()),
			HttpMethod:          http.MethodGet,
			MockMethodName:      GetTaskMethod,
			ReturnedData:        &TestTask,
			ReturnedError:       fmt.Errorf("failed load task by id"),
			ExpectedCalledTimes: 1,
			ExpectedStatusCode:  http.StatusInternalServerError,
		},
	}

	t.Run("Load task by id", func(t *testing.T) {
		for index, testCase := range loadTaskByIDTestCases {
			testCaseName := fmt.Sprintf("Load task by id case %d", index)
			t.Run(testCaseName, func(t *testing.T) {
				testEnv := common.InitTestAppEnvironment()
				appServer, err := testEnv.BuildAppServer(servConfig)
				assert.NoError(t, err, "failed to build app server")

				testEnv.TaskStorage.
					On(testCase.MockMethodName, matchedBucketID, matchedTaskID).
					Return(testCase.ReturnedData, testCase.ReturnedError)

				req := httptest.NewRequest(testCase.HttpMethod, testCase.TargetURL, nil)

				resp, respErr := appServer.Server.Test(req, -1)
				assert.NoError(t, respErr, "failed to load task by id")
				assert.Equal(t, testCase.ExpectedStatusCode, resp.StatusCode, "unexpected http status code")

				testEnv.TaskStorage.AssertNumberOfCalls(t, testCase.MockMethodName, testCase.ExpectedCalledTimes)
			})
		}
	})
}
