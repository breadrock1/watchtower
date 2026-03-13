package httpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"watchtower/cmd/watchtower/httpserver/form"
	"watchtower/internal/core/cloud/domain"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) CreateStorageObjectsGroup(group fiber.Router) {
	group.Post("/cloud/:bucket/files", s.GetFiles)
	group.Post("/cloud/:bucket/file/copy", s.CopyFile)
	group.Post("/cloud/:bucket/file/move", s.MoveFile)
	group.Put("/cloud/:bucket/file/upload", s.UploadFile)
	group.Post("/cloud/:bucket/file/download", s.DownloadFile)
	group.Delete("/cloud/:bucket/file", s.RemoveFile2)
	group.Delete("/cloud/:bucket/file/remove", s.RemoveFile)
	group.Post("/cloud/:bucket/file/attributes", s.GetFileInfo)
	group.Post("/cloud/:bucket/file/share", s.ShareFile)
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
// @Success 200 {object} form.Success "Ok"
// @Failure	400 {object} form.BadRequestError "Bad Request error"
// @Failure	404 {object} form.NotFoundError "Bucket or file not found"
// @Failure	500 {object} form.InternalServerError "Internal server error"
// @Failure	503 {object} form.ServerUnavailableError "Server does not available"
// @Router /api/v1/cloud/{bucket}/file/copy [post]
func (s *Server) CopyFile(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()
	bucket := eCtx.Params("bucket")

	var jsonForm form.CopyFileForm
	err := json.Unmarshal(eCtx.Body(), &jsonForm)
	if err != nil {
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	params := &domain.CopyObjectParams{
		SourcePath:      jsonForm.SrcPath,
		DestinationPath: jsonForm.DstPath,
	}

	err = s.state.GetObjectStorage().CopyObject(ctx, bucket, params)
	if err != nil {
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return eCtx.Status(fiber.StatusOK).SendString("Ok")
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
// @Success 200 {object} form.Success "Ok"
// @Failure	400 {object} form.BadRequestError "Bad Request error"
// @Failure	404 {object} form.NotFoundError "Bucket or file not found"
// @Failure	500 {object} form.InternalServerError "Internal server error"
// @Failure	503 {object} form.ServerUnavailableError "Server does not available"
// @Router /api/v1/cloud/{bucket}/file/move [post]
func (s *Server) MoveFile(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()
	bucket := eCtx.Params("bucket")

	var jsonForm form.CopyFileForm
	err := json.Unmarshal(eCtx.Body(), &jsonForm)
	if err != nil {
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	params := &domain.CopyObjectParams{
		SourcePath:      jsonForm.SrcPath,
		DestinationPath: jsonForm.DstPath,
	}

	err = s.state.GetObjectStorage().MoveObject(ctx, bucket, params)
	if err != nil {
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return eCtx.Status(fiber.StatusOK).SendString("Ok")
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
// @Success 200 {object} form.Success "Ok"
// @Failure	400 {object} form.BadRequestError "Bad Request error"
// @Failure	404 {object} form.NotFoundError "Bucket not found"
// @Failure	500 {object} form.InternalServerError "Internal server error"
// @Failure	503 {object} form.ServerUnavailableError "Server does not available"
// @Router /api/v1/cloud/{bucket}/file/upload [put]
func (s *Server) UploadFile(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()

	var fileData bytes.Buffer

	multipartForm, err := eCtx.MultipartForm()
	if err != nil {
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	bucket := eCtx.Params("bucket")
	exist, err := s.state.GetObjectStorage().IsBucketExists(ctx, bucket)
	if err != nil {
		return eCtx.Status(http.StatusBadRequest).SendString(err.Error())
	}

	if !exist {
		err = fmt.Errorf("specified bucket %s does not exist", bucket)
		return eCtx.Status(http.StatusNotFound).SendString(err.Error())
	}

	if multipartForm.File["files"] == nil {
		err = fmt.Errorf("there are no files into multipart form")
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	expired := eCtx.Query("expired")
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

		params := &domain.UploadObjectParams{
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

	return eCtx.Status(fiber.StatusOK).JSON(uploadedFiles)
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
// @Success 200 {file} io.Writer "Returned file bytes"
// @Failure	400 {object} form.BadRequestError "Bad Request error"
// @Failure	404 {object} form.NotFoundError "Bucket not found"
// @Failure	500 {object} form.InternalServerError "Internal server error"
// @Failure	503 {object} form.ServerUnavailableError "Server does not available"
// @Router /api/v1/cloud/{bucket}/file/download [post]
func (s *Server) DownloadFile(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()
	bucket := eCtx.Params("bucket")

	var jsonForm form.DownloadFileForm
	err := json.Unmarshal(eCtx.Body(), &jsonForm)
	if err != nil {
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	fileData, err := s.state.GetObjectStorage().GetObjectData(ctx, bucket, jsonForm.FileName)
	if err != nil {
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	defer fileData.Reset()

	return eCtx.Send(fileData.Bytes())
}

// RemoveFile
// @Summary Remove file from cloud
// @Description Remove file from cloud
// @ID remove-file
// @Tags files
// @Produce  json
// @Param bucket path string true "Bucket name to remove file"
// @Param jsonQuery body form.RemoveFileForm true "Parameters to remove file"
// @Success 200 {object} form.Success "Ok"
// @Failure	400 {object} form.BadRequestError "Bad Request error"
// @Failure	404 {object} form.NotFoundError "Bucket not found"
// @Failure	500 {object} form.InternalServerError "Internal server error"
// @Failure	503 {object} form.ServerUnavailableError "Server does not available"
// @Router /api/v1/cloud/{bucket}/file/remove [delete]
func (s *Server) RemoveFile(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()
	bucket := eCtx.Params("bucket")

	var jsonForm form.RemoveFileForm
	err := json.Unmarshal(eCtx.Body(), &jsonForm)
	if err != nil {
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if err := s.state.GetObjectStorage().DeleteObject(ctx, bucket, jsonForm.FileName); err != nil {
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return eCtx.Status(fiber.StatusOK).SendString("Ok")
}

// RemoveFile2
// @Summary Remove file from cloud
// @Description Remove file from cloud
// @ID remove-file-2
// @Tags files
// @Produce  json
// @Param bucket path string true "Bucket name to remove file"
// @Param file_name query string true "Parameters to remove file"
// @Success 200 {object} form.Success "Ok"
// @Failure	400 {object} form.BadRequestError "Bad Request error"
// @Failure	404 {object} form.NotFoundError "Bucket not found"
// @Failure	500 {object} form.InternalServerError "Internal server error"
// @Failure	503 {object} form.ServerUnavailableError "Server does not available"
// @Router /api/v1/cloud/{bucket}/file [delete]
func (s *Server) RemoveFile2(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()
	bucket := eCtx.Params("bucket")
	fileName := eCtx.Query("file_name")
	if err := s.state.GetObjectStorage().DeleteObject(ctx, bucket, fileName); err != nil {
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return eCtx.Status(fiber.StatusOK).SendString("Ok")
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
// @Success 200 {object} form.Success "Ok"
// @Failure	400 {object} form.BadRequestError "Bad Request error"
// @Failure	404 {object} form.NotFoundError "Bucket not found"
// @Failure	500 {object} form.InternalServerError "Internal server error"
// @Failure	503 {object} form.ServerUnavailableError "Server does not available"
// @Router /api/v1/cloud/{bucket}/files [post]
func (s *Server) GetFiles(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()
	bucket := eCtx.Params("bucket")

	var jsonForm form.GetFilesForm
	err := json.Unmarshal(eCtx.Body(), &jsonForm)
	if err != nil {
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	params := &domain.GetObjectsParams{
		PrefixPath: jsonForm.DirectoryName,
	}

	listObjects, err := s.state.GetObjectStorage().LoadBucketObjects(ctx, bucket, params)
	if err != nil {
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	objectsDto := make([]form.ObjectSchema, len(listObjects))
	for index, object := range listObjects {
		objectsDto[index] = form.ObjectFromDomain(object)
	}

	return eCtx.Status(fiber.StatusOK).JSON(objectsDto)
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
// @Success 200 {object} form.Success "Ok"
// @Failure	400 {object} form.BadRequestError "Bad Request error"
// @Failure	404 {object} form.NotFoundError "Bucket or Object not found"
// @Failure	500 {object} form.InternalServerError "Internal server error"
// @Failure	503 {object} form.ServerUnavailableError "Server does not available"
// @Router /api/v1/cloud/{bucket}/file/attributes [post]
func (s *Server) GetFileInfo(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()
	bucket := eCtx.Params("bucket")

	var jsonForm form.GetFileAttributesForm
	err := json.Unmarshal(eCtx.Body(), &jsonForm)
	if err != nil {
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	object, err := s.state.GetObjectStorage().GetObjectInfo(ctx, bucket, jsonForm.FilePath)
	if err != nil {
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	objectSchema := form.ObjectFromDomain(*object)
	return eCtx.Status(fiber.StatusOK).JSON(objectSchema)
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
// @Success 200 {object} form.Success "URL with shared file"
// @Failure	400 {object} form.BadRequestError "Bad Request error"
// @Failure	404 {object} form.NotFoundError "Bucket of object not found"
// @Failure	500 {object} form.InternalServerError "Internal server error"
// @Failure	503 {object} form.ServerUnavailableError "Server does not available"
// @Router /api/v1/cloud/{bucket}/file/share [post]
func (s *Server) ShareFile(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()
	bucket := eCtx.Params("bucket")

	var jsonForm form.ShareFileForm
	err := json.Unmarshal(eCtx.Body(), &jsonForm)
	if err != nil {
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	expired := time.Second * time.Duration(jsonForm.ExpiredSecs)
	params := &domain.ShareObjectParams{FilePath: jsonForm.FilePath, Expired: expired}
	url, err := s.state.GetObjectStorage().GenShareURL(ctx, bucket, params)
	if err != nil {
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return eCtx.Status(fiber.StatusOK).JSON(form.SuccessResponse(url))
}
