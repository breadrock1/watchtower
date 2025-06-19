package httpserver

import (
	"encoding/json"
	"watchtower/internal/application/mapping"

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
func (s *Server) AttachDirectory(eCtx echo.Context) error {
	jsonForm := &AttachDirectoryForm{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	err := decoder.Decode(jsonForm)
	if err != nil {
		return err
	}

	dir := dto.Directory{
		Bucket: jsonForm.BucketName,
		Path:   jsonForm.Directory,
	}

	ctx := eCtx.Request().Context()
	err = s.watcher.AttachWatchedDir(ctx, dir)
	if err != nil {
		return err
	}

	return eCtx.JSON(200, createStatusResponse(200, "Ok"))
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
func (s *Server) DetachDirectory(eCtx echo.Context) error {
	bucket := eCtx.Param("bucket")

	ctx := eCtx.Request().Context()
	if err := s.watcher.DetachWatchedDir(ctx, bucket); err != nil {
		return err
	}

	return eCtx.JSON(200, createStatusResponse(200, "Ok"))
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
func (s *Server) FetchDocumentsByStatus(eCtx echo.Context) error {
	jsonForm := &FetchDocumentsList{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	if err := decoder.Decode(jsonForm); err != nil {
		return err
	}

	ctx := eCtx.Request().Context()
	tasks, err := s.taskManager.GetAll(ctx, jsonForm.BucketName)
	if err != nil {
		return err
	}

	foundedTasks := make([]*dto.TaskEvent, 0)
	inputTaskStatus := mapping.TaskStatusFromString(jsonForm.Status)
	for _, task := range tasks {
		if task.Status == mapping.TaskStatusToInt(inputTaskStatus) {
			foundedTasks = append(foundedTasks, task)
		}
	}

	return eCtx.JSON(200, foundedTasks)
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
func (s *Server) GetAllProcessingDocuments(eCtx echo.Context) error {
	jsonForm := &FetchAllDocuments{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	if err := decoder.Decode(jsonForm); err != nil {
		return err
	}

	ctx := eCtx.Request().Context()
	documents, err := s.taskManager.GetAll(ctx, jsonForm.BucketName)
	if err != nil {
		return err
	}

	return eCtx.JSON(200, documents)
}
