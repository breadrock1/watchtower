package httpserver

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/exp/slices"
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
	ctx := eCtx.Request().Context()
	ctx, span := s.tracer.Start(ctx, "load-tasks")
	defer span.End()

	bucket := eCtx.Param("bucket")
	if bucket == "" {
		err := fmt.Errorf("bucket parameter is required")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "bucket is required")
	}

	tasks, err := s.uc.GetTaskManager().GetAll(ctx, bucket)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	status := eCtx.QueryParam("status")
	if status == "" {
		return eCtx.JSON(200, tasks)
	}

	inputTaskStatus, err := mapping.TaskStatusFromString(status)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
	ctx := eCtx.Request().Context()
	ctx, span := s.tracer.Start(ctx, "load-task-by-id")
	defer span.End()

	bucket := eCtx.Param("bucket")
	if bucket == "" {
		err := fmt.Errorf("bucket parameter is required")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "bucket is required")
	}

	taskID := eCtx.Param("task_id")
	if taskID == "" {
		err := fmt.Errorf("task-id parameter is required")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "task_id is required")
	}

	task, err := s.uc.GetTaskManager().Get(ctx, bucket, taskID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return eCtx.JSON(200, task)
}
