package httpserver

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slices"
	"watchtower/internal/domain/core/process"
)

func (s *Server) CreateTasksGroup() error {
	group := s.server.Group("/api/v1/tasks")

	group.GET("/:bucket", s.LoadTasks)
	group.GET("/:bucket/:task_id", s.LoadTaskByID)

	return nil
}

// LoadTasks
// @Summary Load processing tasks of uploaded files into bucket
// @Description Load tasks (processing/unrecognized/done) of uploaded files
// @ID load-tasks
// @Tags tasks
// @Accept  json
// @Produce json
// @Param bucket path string true "Name id of uploaded files"
// @Param status query string false "Status tasks to filter target result"
// @Success 200 {object} []dto.Task "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /tasks/{bucket} [get]
func (s *Server) LoadTasks(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	bucket := eCtx.Param("bucket")
	if bucket == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "bucket is required")
	}

	tasks, err := s.state.GetTaskProcessor().GetBucketTasks(ctx, bucket)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	status := eCtx.QueryParam("status")
	if status == "" {
		return eCtx.JSON(200, tasks)
	}

	inputTaskStatus, err := strconv.Atoi(status)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "unknown status")
	}

	taskStatus := process.TaskStatus(inputTaskStatus)
	foundedTasks := slices.DeleteFunc(tasks, func(task *process.Task) bool {
		return task.Status != taskStatus
	})

	return eCtx.JSON(200, foundedTasks)
}

// LoadTaskByID
// @Summary Load processing task by id
// @Description Load processing/unrecognized/done task by id of uploaded file
// @ID load-task-by-id
// @Tags tasks
// @Accept  json
// @Produce json
// @Param bucket path string true "Name id of processing task"
// @Param task_id path string true "Task ID"
// @Success 200 {object} dto.Task "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /tasks/{bucket}/{task_id} [get]
func (s *Server) LoadTaskByID(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	bucket := eCtx.Param("bucket")
	if bucket == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "bucket is required")
	}

	taskIDParam := eCtx.Param("task_id")
	if taskIDParam == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "task_id is required")
	}

	taskID, err := uuid.Parse(taskIDParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	task, err := s.state.GetTaskProcessor().GetTask(ctx, bucket, taskID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return eCtx.JSON(200, task)
}
