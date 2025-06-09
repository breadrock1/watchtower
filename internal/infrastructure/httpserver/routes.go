package httpserver

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
	"watchtower/internal/application/dto"
)

func (s *Server) CreateWatcherGroup() error {
	group := s.server.Group("/watcher")

	group.POST("/attach", s.AttachDirectory)
	group.DELETE("/:bucket", s.DetachDirectory)

	return nil
}

func (s *Server) CreateTasksGroup() error {
	group := s.server.Group("/tasks")

	group.POST("/fetch", s.FetchDocumentsByStatus)
	group.POST("/all", s.GetAllProcessingDocuments)

	return nil
}

// AttachDirectory
// @Summary Attach new directory to watcher
// @Description Attach new directory to watcher
// @ID folders-attach
// @Tags watcher
// @Accept  json
// @Produce json
// @Param jsonQuery body AttachDirectoryForm true "File entity"
// @Success 200 {object} ResponseForm "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /watcher/attach [post]
func (s *Server) AttachDirectory(c echo.Context) error {
	jsonForm := &AttachDirectoryForm{}
	decoder := json.NewDecoder(c.Request().Body)
	err := decoder.Decode(jsonForm)
	if err != nil {
		return err
	}

	dir := dto.Directory{
		Bucket: jsonForm.BucketName,
		Path:   jsonForm.Directory,
	}

	ctx := c.Request().Context()
	err = s.watcher.AttachWatchedDir(ctx, dir)
	if err != nil {
		return err
	}

	return c.JSON(200, createStatusResponse(200, "Ok"))
}

// DetachDirectory
// @Summary Attach new directory to watcher
// @Description Attach new directory to watcher
// @ID folders-detach
// @Tags watcher
// @Accept  json
// @Produce json
// @Param bucket path string true "Folder ids"
// @Success 200 {object} ResponseForm "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /watcher/{bucket} [delete]
func (s *Server) DetachDirectory(c echo.Context) error {
	bucket := c.Param("bucket")

	ctx := c.Request().Context()
	if err := s.watcher.DetachWatchedDir(ctx, bucket); err != nil {
		return err
	}

	return c.JSON(200, createStatusResponse(200, "Ok"))
}

// FetchDocumentsByStatus
// @Summary Fetch processing documents
// @Description Load processing/unrecognized/done documents by names list
// @ID fetch-documents
// @Tags tasks
// @Accept  json
// @Produce json
// @Param jsonQuery body FetchDocumentsList true "File names to fetch processing status"
// @Success 200 {object} []dto.TaskEvent "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /tasks/fetch [post]
func (s *Server) FetchDocumentsByStatus(c echo.Context) error {
	jsonForm := &FetchDocumentsList{}
	decoder := json.NewDecoder(c.Request().Body)
	if err := decoder.Decode(jsonForm); err != nil {
		return err
	}

	ctx := c.Request().Context()
	tasks, err := s.taskManager.GetAll(ctx, jsonForm.BucketName)
	if err != nil {
		return err
	}

	foundedTasks := make([]*dto.TaskEvent, 0)
	//json
	inputTaskStatus := dto.TaskStatusFromString(jsonForm.Status)
	for _, task := range tasks {
		if task.Status == inputTaskStatus {
			foundedTasks = append(foundedTasks, task)
		}
	}

	return c.JSON(200, foundedTasks)
}

// GetAllProcessingDocuments
// @Summary Get all processing documents
// @Description Get all processing documents
// @ID get-processing-documents
// @Tags tasks
// @Accept  json
// @Param jsonQuery body FetchAllDocuments true "File names to fetch processing status"
// @Success 200 {object} ResponseForm "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /tasks/all [post]
func (s *Server) GetAllProcessingDocuments(c echo.Context) error {
	jsonForm := &FetchAllDocuments{}
	decoder := json.NewDecoder(c.Request().Body)
	if err := decoder.Decode(jsonForm); err != nil {
		return err
	}

	ctx := c.Request().Context()
	documents, err := s.taskManager.GetAll(ctx, jsonForm.BucketName)
	if err != nil {
		return err
	}

	return c.JSON(200, documents)
}
