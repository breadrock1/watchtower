package httpserver

import (
	"golang.org/x/exp/slices"
	"net/http"

	"github.com/labstack/echo/v4"
	"watchtower/internal/application/dto"
	"watchtower/internal/application/mapping"
)

func (s *Server) CreateTasksGroup() error {
	group := s.server.Group("/tasks")

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
// @Param bucket path string true "Bucket id of uploaded files"
// @Param status query string false "Status tasks to filter target result"
// @Success 200 {object} []dto.TaskEvent "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /tasks/{bucket} [get]
func (s *Server) LoadTasks(eCtx echo.Context) error {
	bucket := eCtx.Param("bucket")
	if bucket == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "bucket is required")
	}

	ctx := eCtx.Request().Context()
	tCtx, span := GetTracer().Start(ctx, "load-buckets")
	defer span.End()

	tasks, err := s.uc.GetTaskManager().GetAll(tCtx, bucket)
	if err != nil {
		span.RecordError(err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	status := eCtx.QueryParam("status")
	if status == "" {
		return eCtx.JSON(200, tasks)
	}

	inputTaskStatus, err := mapping.TaskStatusFromString(status)
	if err != nil {
		span.RecordError(err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "unknown status")
	}

	foundedTasks := slices.DeleteFunc(tasks, func(event *dto.TaskEvent) bool {
		return event.Status != inputTaskStatus
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
// @Param bucket path string true "Bucket id of processing task"
// @Param task_id path string true "Task ID"
// @Success 200 {object} dto.TaskEvent "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /tasks/{bucket}/{task_id} [get]
func (s *Server) LoadTaskByID(eCtx echo.Context) error {
	bucket := eCtx.Param("bucket")
	if bucket == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "bucket is required")
	}

	taskID := eCtx.Param("task_id")
	if taskID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "task_id is required")
	}

	ctx := eCtx.Request().Context()
	task, err := s.uc.GetTaskManager().Get(ctx, bucket, taskID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return eCtx.JSON(200, task)
}
