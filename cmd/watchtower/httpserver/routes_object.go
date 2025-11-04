package httpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"watchtower/cmd/watchtower/httpserver/form"
	"watchtower/internal/core/cloud/domain"
)

func (s *Server) CreateStorageObjectsGroup() error {
	group := s.server.Group("/api/v1/cloud")

	group.POST("/:bucket/files", s.GetFiles)
	group.POST("/:bucket/file/copy", s.CopyFile)
	group.POST("/:bucket/file/move", s.MoveFile)
	group.PUT("/:bucket/file/upload", s.UploadFile)
	group.POST("/:bucket/file/download", s.DownloadFile)
	group.DELETE("/:bucket/file", s.RemoveFile2)
	group.DELETE("/:bucket/file/remove", s.RemoveFile)
	group.POST("/:bucket/file/attributes", s.GetFileInfo)

	group.POST("/:bucket/file/share", s.ShareFile)

	return nil
}

// CopyFile
// @Summary Copy file to another location into bucket
// @Description Copy file to another location into bucket
// @ID copy-file
// @Tags files
// @Accept  json
// @Produce json
// @Param bucket path string true "Bucket name of src file"
// @Param jsonQuery body form.CopyFileForm true "Params to copy file"
// @Success 200 {object} form.ResponseForm "Ok"
// @Failure	400 {object} form.BadRequestForm "Bad Request message"
// @Failure	503 {object} form.ServerErrorForm "Server does not available"
// @Router /cloud/{bucket}/file/copy [post]
func (s *Server) CopyFile(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	bucket := eCtx.Param("bucket")

	jsonForm := &form.CopyFileForm{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	err := decoder.Decode(jsonForm)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	params := domain.CopyObjectParams{
		SourcePath:      jsonForm.SrcPath,
		DestinationPath: jsonForm.DstPath,
	}

	err = s.state.GetObjectStorage().CopyObject(ctx, bucket, params)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return eCtx.JSON(200, form.CreateStatusResponse("Ok"))
}

// MoveFile
// @Summary Move file to another location into bucket
// @Description Move file to another location into bucket
// @ID move-file
// @Tags files
// @Accept  json
// @Produce json
// @Param bucket path string true "Bucket name of src file"
// @Param jsonQuery body form.CopyFileForm true "Params to move file"
// @Success 200 {object} form.ResponseForm "Ok"
// @Failure	400 {object} form.BadRequestForm "Bad Request message"
// @Failure	503 {object} form.ServerErrorForm "Server does not available"
// @Router /cloud/{bucket}/file/move [post]
func (s *Server) MoveFile(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	bucket := eCtx.Param("bucket")

	jsonForm := &form.CopyFileForm{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	err := decoder.Decode(jsonForm)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	params := domain.CopyObjectParams{
		SourcePath:      jsonForm.SrcPath,
		DestinationPath: jsonForm.DstPath,
	}

	err = s.state.GetObjectStorage().MoveObject(ctx, bucket, params)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return eCtx.JSON(200, form.CreateStatusResponse("Ok"))
}

// UploadFile
// @Summary Upload files to cloud
// @Description Upload files to cloud
// @ID upload-files
// @Tags files
// @Accept  multipart/form
// @Produce  json
// @Param bucket path string true "Bucket name to upload files"
// @Param files formData file true "Files multipart form"
// @Param expired query string false "File datetime expired like 2025-01-01T12:01:01Z"
// @Success 200 {object} form.ResponseForm "Ok"
// @Failure	400 {object} form.BadRequestForm "Bad Request message"
// @Failure	503 {object} form.ServerErrorForm "Server does not available"
// @Router /cloud/{bucket}/file/upload [put]
func (s *Server) UploadFile(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()

	var fileData bytes.Buffer

	multipartForm, err := eCtx.MultipartForm()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	bucket := eCtx.Param("bucket")
	exist, err := s.state.GetObjectStorage().IsBucketExists(eCtx.Request().Context(), bucket)
	if err != nil || !exist {
		retErr := fmt.Errorf("specified bucket %s does not exist", bucket)
		return echo.NewHTTPError(http.StatusBadRequest, retErr.Error())
	}

	if multipartForm.File["files"] == nil {
		err = fmt.Errorf("there are no files into multipart form")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	expired := eCtx.QueryParam("expired")
	timeVal, timeParseErr := time.Parse(time.RFC3339, expired)
	if timeParseErr != nil {
		slog.Warn("failed to parse expired time param",
			slog.String("err", timeParseErr.Error()),
		)
	}

	uploadedFiles := make([]form.TaskSchema, len(multipartForm.File["files"]))
	for index, fileForm := range multipartForm.File["files"] {
		fileName := fileForm.Filename
		fileHandler, err := fileForm.Open()
		if err != nil {
			slog.Error("failed to open file form", slog.String("err", err.Error()))
			continue
		}
		defer func() {
			if err := fileHandler.Close(); err != nil {
				slog.Error("failed to close file handler",
					slog.String("file", fileName),
					slog.String("err", err.Error()),
				)
				return
			}
		}()

		fileData.Reset()
		_, err = fileData.ReadFrom(fileHandler)
		if err != nil {
			slog.Error("failed to read file form",
				slog.String("file", fileName),
				slog.String("err", err.Error()),
			)
			continue
		}

		params := domain.UploadObjectParams{
			FilePath: fileName,
			FileData: &fileData,
			Expired:  &timeVal,
		}

		task, err := s.state.UploadFile(ctx, bucket, params)
		if err != nil {
			slog.Error("failed to upload file to cloud",
				slog.String("file", fileName),
				slog.String("err", err.Error()),
			)
			continue
		}

		uploadedFiles[index] = form.TaskFromDomain(*task)
	}

	return eCtx.JSON(200, uploadedFiles)
}

// DownloadFile
// @Summary Download file from cloud
// @Description Download file from cloud
// @ID download-file
// @Tags files
// @Accept  json
// @Produce json
// @Param bucket path string true "Bucket name to download file"
// @Param jsonQuery body form.DownloadFileForm true "Parameters to download file"
// @Success 200 {file} io.Writer "Ok"
// @Failure	400 {object} form.BadRequestForm "Bad Request message"
// @Failure	503 {object} form.ServerErrorForm "Server does not available"
// @Router /cloud/{bucket}/file/download [post]
func (s *Server) DownloadFile(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	bucket := eCtx.Param("bucket")

	jsonForm := &form.DownloadFileForm{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	if err := decoder.Decode(jsonForm); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	fileData, err := s.state.GetObjectStorage().GetObjectData(ctx, bucket, jsonForm.FileName)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	defer fileData.Reset()

	return eCtx.Blob(200, echo.MIMEMultipartForm, fileData.Bytes())
}

// RemoveFile
// @Summary Remove file from cloud
// @Description Remove file from cloud
// @ID remove-file
// @Tags files
// @Produce  json
// @Param bucket path string true "Bucket name to remove file"
// @Param jsonQuery body form.RemoveFileForm true "Parameters to remove file"
// @Success 200 {object} form.ResponseForm "Ok"
// @Failure	400 {object} form.BadRequestForm "Bad Request message"
// @Failure	503 {object} form.ServerErrorForm "Server does not available"
// @Router /cloud/{bucket}/file/remove [delete]
func (s *Server) RemoveFile(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	bucket := eCtx.Param("bucket")

	jsonForm := &form.RemoveFileForm{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	if err := decoder.Decode(jsonForm); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := s.state.GetObjectStorage().DeleteObject(ctx, bucket, jsonForm.FileName); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return eCtx.JSON(200, form.CreateStatusResponse("Ok"))
}

// RemoveFile2
// @Summary Remove file from cloud
// @Description Remove file from cloud
// @ID remove-file-2
// @Tags files
// @Produce  json
// @Param bucket path string true "Bucket name to remove file"
// @Param file_name query string true "Parameters to remove file"
// @Success 200 {object} form.ResponseForm "Ok"
// @Failure	400 {object} form.BadRequestForm "Bad Request message"
// @Failure	503 {object} form.ServerErrorForm "Server does not available"
// @Router /cloud/{bucket}/file [delete]
func (s *Server) RemoveFile2(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	bucket := eCtx.Param("bucket")
	fileName := eCtx.QueryParam("file_name")
	if err := s.state.GetObjectStorage().DeleteObject(ctx, bucket, fileName); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return eCtx.JSON(200, form.CreateStatusResponse("Ok"))
}

// GetFiles
// @Summary Get files list into bucket
// @Description Get files list into bucket
// @ID get-list-files
// @Tags files
// @Accept  json
// @Produce json
// @Param bucket path string true "Bucket name to get list files"
// @Param jsonQuery body form.GetFilesForm true "Parameters to get list files"
// @Success 200 {object} form.ResponseForm "Ok"
// @Failure	400 {object} form.BadRequestForm "Bad Request message"
// @Failure	503 {object} form.ServerErrorForm "Server does not available"
// @Router /cloud/{bucket}/files [post]
func (s *Server) GetFiles(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	bucket := eCtx.Param("bucket")

	jsonForm := &form.GetFilesForm{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	err := decoder.Decode(jsonForm)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	params := domain.GetObjectsParams{
		PrefixPath: jsonForm.DirectoryName,
	}

	listObjects, err := s.state.GetObjectStorage().GetBucketObjects(ctx, bucket, params)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	objectsDto := make([]form.ObjectSchema, len(listObjects))
	for index, object := range listObjects {
		objectsDto[index] = form.ObjectFromDomain(object)
	}

	return eCtx.JSON(200, objectsDto)
}

// GetFileInfo
// @Summary Get file attributes
// @Description Get file attributes
// @ID get-file-attrs
// @Tags files
// @Accept  json
// @Produce json
// @Param bucket path string true "Bucket name to get list files"
// @Param jsonQuery body form.GetFileAttributesForm true "Parameters to get list files"
// @Success 200 {object} form.ResponseForm "Ok"
// @Failure	400 {object} form.BadRequestForm "Bad Request message"
// @Failure	503 {object} form.ServerErrorForm "Server does not available"
// @Router /cloud/{bucket}/file/attributes [post]
func (s *Server) GetFileInfo(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	bucket := eCtx.Param("bucket")

	jsonForm := &form.GetFileAttributesForm{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	err := decoder.Decode(jsonForm)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	object, err := s.state.GetObjectStorage().GetObjectInfo(ctx, bucket, jsonForm.FilePath)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return eCtx.JSON(200, form.ObjectFromDomain(*object))
}

// ShareFile
// @Summary Get share URL for file
// @Description Get share URL for file
// @ID share-file
// @Tags share
// @Accept  json
// @Produce json
// @Param bucket path string true "Bucket name to share file"
// @Param jsonQuery body form.ShareFileForm true "Parameters to share file"
// @Success 200 {object} form.ResponseForm "Ok"
// @Failure	400 {object} form.BadRequestForm "Bad Request message"
// @Failure	503 {object} form.ServerErrorForm "Server does not available"
// @Router /cloud/{bucket}/file/share [post]
func (s *Server) ShareFile(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	bucket := eCtx.Param("bucket")

	shareForm := &form.ShareFileForm{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	err := decoder.Decode(shareForm)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	expired := time.Second * time.Duration(shareForm.ExpiredSecs)
	params := domain.ShareObjectParams{FilePath: shareForm.FilePath, Expired: expired}
	url, err := s.state.GetObjectStorage().GenShareURL(ctx, bucket, params)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return eCtx.JSON(200, form.CreateStatusResponse(url))
}
