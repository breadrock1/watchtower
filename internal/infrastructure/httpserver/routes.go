package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"watchtower/internal/application/dto"
	"watchtower/internal/application/mapping"
)

func (s *Server) CreateWatcherGroup() error {
	group := s.server.Group("/watcher")

	group.POST("/:bucket", s.AddDirectoryToWatcher)
	group.DELETE("/:bucket", s.DeleteDirectoryFromWatcher)

	return nil
}

func (s *Server) CreateTasksGroup() error {
	group := s.server.Group("/tasks")

	group.GET("/:bucket/all", s.LoadTasks)
	group.GET("/:bucket", s.GetTask)

	return nil
}

// AddDirectoryToWatcher
// @Summary Attach new directory to watcher
// @Description Attach new directory to watcher
// @ID folders-attach
// @Tags watcher
// @Accept  json
// @Produce json
// @Param jsonQuery body AddDirectoryToWatcherForm true "Bucket form"
// @Success 200 {object} ResponseForm "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /watcher/{bucket} [post]
func (s *Server) AddDirectoryToWatcher(eCtx echo.Context) error {
	jsonForm := &AddDirectoryToWatcherForm{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	err := decoder.Decode(jsonForm)
	if err != nil {
		return err
	}

	dir := dto.Directory{
		Bucket: jsonForm.BucketName,
		Path:   jsonForm.Suffix,
	}

	ctx := eCtx.Request().Context()
	err = s.watcher.AttachWatchedDir(ctx, dir)
	if err != nil {
		return err
	}

	return eCtx.JSON(201, createStatusResponse(200, "Ok"))
}

// DeleteDirectoryFromWatcher
// @Summary Attach new directory to watcher
// @Description Attach new directory to watcher
// @ID folders-detach
// @Tags watcher
// @Accept  json
// @Produce json
// @Param bucket path string true "Bucket ids"
// @Success 200 {object} ResponseForm "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /watcher/{bucket} [delete]
func (s *Server) DeleteDirectoryFromWatcher(eCtx echo.Context) error {
	bucket := eCtx.Param("bucket")
	if bucket == "" {
		return echo.NewHTTPError(http.StatusNotFound, "unknown bucket")
	}

	ctx := eCtx.Request().Context()
	if err := s.watcher.DetachWatchedDir(ctx, bucket); err != nil {
		return err
	}

	return eCtx.JSON(200, createStatusResponse(200, "Ok"))
}

// LoadTasks
// @Summary Load tasks of processing documents
// @Description Load tasks of processing/unrecognized/done documents
// @ID load-tasks
// @Tags tasks
// @Accept  json
// @Produce json
// @Param bucket path string true "Bucket id"
// @Param status query string false "Status"
// @Success 200 {object} []dto.TaskEvent "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /tasks/{bucket}/all [get]
func (s *Server) LoadTasks(eCtx echo.Context) error {
	bucket := eCtx.Param("bucket")
	if bucket == "" {
		return echo.NewHTTPError(http.StatusNotFound, "unknown bucket")
	}

	ctx := eCtx.Request().Context()
	tasks, err := s.taskManager.GetAll(ctx, bucket)
	if err != nil {
		return err
	}

	status := eCtx.QueryParam("status")
	if status == "" {
		return eCtx.JSON(200, tasks)
	}

	inputTaskStatus, err := mapping.TaskStatusFromString(status)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "unknown status")
	}

	foundedTasks := make([]*dto.TaskEvent, 0)
	for _, task := range tasks {
		if task.Status == inputTaskStatus {
			foundedTasks = append(foundedTasks, task)
		}
	}

	return eCtx.JSON(200, foundedTasks)
}

// GetTask
// @Summary Get processing task
// @Description Get processing/unrecognized/done task document
// @ID get-task
// @Tags tasks
// @Accept  json
// @Produce json
// @Param bucket path string true "Bucket id"
// @Param file query string true "File path into bucket"
// @Success 200 {object} dto.TaskEvent "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /tasks/{bucket} [get]
func (s *Server) GetTask(eCtx echo.Context) error {
	bucket := eCtx.Param("bucket")
	if bucket == "" {
		return echo.NewHTTPError(http.StatusNotFound, "unknown bucket")
	}

	filePath := eCtx.Param("file")

	ctx := eCtx.Request().Context()
	task, err := s.taskManager.Get(ctx, bucket, filePath)
	if err != nil {
		return err
	}

	return eCtx.JSON(200, task)
}
