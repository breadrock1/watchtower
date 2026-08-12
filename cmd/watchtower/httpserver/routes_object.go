package httpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"watchtower/cmd/watchtower/httpserver/form"
	"watchtower/cmd/watchtower/httpserver/mw"
	"watchtower/internal/core/cloud/domain"
)

const FolderFileKeeper = ".keeper"

func (s *Server) CreateStorageObjectsGroup(group fiber.Router) {
	group.Post("/cloud/:bucket/files", s.GetFiles)
	group.Patch("/cloud/:bucket/file", s.CopyFile)
	group.Put("/cloud/:bucket/file/upload", s.UploadFile)
	group.Post("/cloud/:bucket/file/download", s.DownloadFile)
	group.Post("/cloud/:bucket/folder", s.CreateFolder)
	group.Delete("/cloud/:bucket/folder", s.DeleteFolder)
	group.Delete("/cloud/:bucket/file", s.RemoveFile2)
	group.Delete("/cloud/:bucket/file/remove", s.RemoveFile)
	group.Post("/cloud/:bucket/file/attributes", s.GetFileInfo)
	group.Post("/cloud/:bucket/file/share", s.ShareFile)
}

// CreateFolder
// @Summary Create empty folder into cloud storage
// @Description Create empty folder into cloud storage
// @ID create-folder
// @Tags files
// @Accept  application/json
// @Produce  json
// @Param X-Organization-Id header string false "Unique Organization ID to choose s3 instance"
// @Param bucket path string true "Bucket name to create folder"
// @Param jsonQuery body form.FolderForm true "Params to create folder"
// @Success 200 {object} form.Success "Ok"
// @Failure	400 {object} form.BadRequestError "Bad Request error"
// @Failure	404 {object} form.NotFoundError "Bucket not found"
// @Failure	500 {object} form.InternalServerError "Internal server error"
// @Failure	503 {object} form.ServerUnavailableError "Server does not available"
// @Router /api/v1/cloud/{bucket}/folder [post]
func (s *Server) CreateFolder(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()

	span := trace.SpanFromContext(ctx)

	bucket, err := ExtractBucketParameter(eCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	span.SetAttributes(attribute.String("bucket", bucket))

	orgID := eCtx.Locals(mw.OrganizationIDKey).(string)
	objStorage, err := s.state.GetObjectStorage(orgID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	exist, err := objStorage.IsBucketExists(ctx, bucket)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(http.StatusBadRequest).SendString(err.Error())
	}

	if !exist {
		err = fmt.Errorf("specified bucket %s does not exist", bucket)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(http.StatusNotFound).SendString(err.Error())
	}

	var jsonForm form.FolderForm
	err = json.Unmarshal(eCtx.Body(), &jsonForm)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	keepFilePath := path.Join(jsonForm.Prefix, FolderFileKeeper)
	params := &domain.UploadObjectParams{
		FilePath: keepFilePath,
		FileData: bytes.NewBufferString(""),
		Expired:  nil,
	}

	_, err = objStorage.StoreObject(ctx, bucket, params)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return eCtx.Status(fiber.StatusCreated).SendString("Ok")
}

// DeleteFolder
// @Summary Delete folder into cloud storage
// @Description Delete empty folder into cloud storage
// @ID delete-folder
// @Tags files
// @Accept  application/json
// @Produce  json
// @Param X-Organization-Id header string false "Unique Organization ID to choose s3 instance"
// @Param bucket path string true "Bucket name to delete folder"
// @Param jsonQuery body form.FolderForm true "Params to delete folder"
// @Success 200 {object} form.Success "Ok"
// @Failure	400 {object} form.BadRequestError "Bad Request error"
// @Failure	404 {object} form.NotFoundError "Bucket not found"
// @Failure	500 {object} form.InternalServerError "Internal server error"
// @Failure	503 {object} form.ServerUnavailableError "Server does not available"
// @Router /api/v1/cloud/{bucket}/folder [delete]
func (s *Server) DeleteFolder(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()

	span := trace.SpanFromContext(ctx)

	bucket, err := ExtractBucketParameter(eCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	span.SetAttributes(attribute.String("bucket", bucket))

	orgID := eCtx.Locals(mw.OrganizationIDKey).(string)
	objStorage, err := s.state.GetObjectStorage(orgID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	exist, err := objStorage.IsBucketExists(ctx, bucket)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(http.StatusBadRequest).SendString(err.Error())
	}

	if !exist {
		err = fmt.Errorf("specified bucket %s does not exist", bucket)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(http.StatusNotFound).SendString(err.Error())
	}

	var jsonForm form.FolderForm
	err = json.Unmarshal(eCtx.Body(), &jsonForm)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	err = objStorage.DeleteObjects(ctx, bucket, jsonForm.Prefix)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
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
// @Param X-Organization-Id header string false "Unique Organization ID to choose s3 instance"
// @Param bucket path string true "Bucket name to upload files"
// @Param prefix formData string false "Prefix to load files"
// @Param files formData file true "Files multipart form"
// @Param expired query string false "File datetime expired like 2025-01-01T12:01:01Z"
// @Param processing query bool false "Allow uploaded file processing"
// @Success 200 {object} form.Success "Ok"
// @Failure	400 {object} form.BadRequestError "Bad Request error"
// @Failure	404 {object} form.NotFoundError "Bucket not found"
// @Failure	500 {object} form.InternalServerError "Internal server error"
// @Failure	503 {object} form.ServerUnavailableError "Server does not available"
// @Router /api/v1/cloud/{bucket}/file/upload [put]
func (s *Server) UploadFile(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()

	span := trace.SpanFromContext(ctx)

	bucket, err := ExtractBucketParameter(eCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	processingFlag, err := ExtractTaskProcessingFlagParameter(eCtx)
	if err != nil {
		slog.Warn("incorrect query parameter", slog.String("err", err.Error()))
	}

	span.SetAttributes(attribute.String("bucket", bucket), attribute.Bool("processing", processingFlag))

	orgID := eCtx.Locals(mw.OrganizationIDKey).(string)
	objStorage, err := s.state.GetObjectStorage(orgID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	exist, err := objStorage.IsBucketExists(ctx, bucket)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(http.StatusBadRequest).SendString(err.Error())
	}

	if !exist {
		err = fmt.Errorf("specified bucket %s does not exist", bucket)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(http.StatusNotFound).SendString(err.Error())
	}

	multipartForm, err := ExtractMultipartForm(eCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	filePrefixParam := ExtractFilePrefixParameter(eCtx)
	expiredDatetime, err := ExtractExpiredDatetime(eCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	uploadedFileNames := make([]string, 0, len(multipartForm.File["files"]))
	for _, fileForm := range multipartForm.File["files"] {
		uploadedFileNames = append(uploadedFileNames, fileForm.Filename)
	}
	eCtx.Locals("file_path", strings.Join(uploadedFileNames, ","))

	var fileData bytes.Buffer
	uploadedFiles := make([]form.TaskSchema, len(multipartForm.File["files"]))
	for index, fileForm := range multipartForm.File["files"] {
		fileName := fileForm.Filename
		filePath := path.Join(filePrefixParam, fileName)
		fileHandler, err := fileForm.Open()
		if err != nil {
			err = fmt.Errorf("failed to open file form: %w", err)
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			slog.Error("multipart error",
				slog.String("file", fileName),
				slog.String("prefix", filePrefixParam),
				slog.String("err", err.Error()),
			)
			continue
		}
		defer func() {
			if err := fileHandler.Close(); err != nil {
				err = fmt.Errorf("failed to close file handler: %w", err)
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
				slog.Error("multipart error",
					slog.String("file", fileName),
					slog.String("prefix", filePrefixParam),
					slog.String("err", err.Error()),
				)
				return
			}
		}()

		fileData.Reset()
		_, err = fileData.ReadFrom(fileHandler)
		if err != nil {
			err = fmt.Errorf("failed to read file form: %w", err)
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			slog.Error("multipart error",
				slog.String("file", fileName),
				slog.String("prefix", filePrefixParam),
				slog.String("err", err.Error()),
			)
			continue
		}

		params := &domain.UploadObjectParams{
			Organization:         orgID,
			FilePath:             filePath,
			FileData:             &fileData,
			Expired:              expiredDatetime,
			CreateProcessingTask: processingFlag,
		}

		task, err := s.state.UploadFile(ctx, bucket, params)
		if err != nil {
			err = fmt.Errorf("failed to upload file form: %w", err)
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			slog.Error("multipart error",
				slog.String("file", fileName),
				slog.String("prefix", filePrefixParam),
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
// @Param X-Organization-Id header string false "Unique Organization ID to choose s3 instance"
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

	span := trace.SpanFromContext(ctx)

	bucket, err := ExtractBucketParameter(eCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	span.SetAttributes(attribute.String("bucket", bucket))

	var jsonForm form.DownloadFileForm
	err = json.Unmarshal(eCtx.Body(), &jsonForm)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	eCtx.Locals("file_path", jsonForm.FileName)

	orgID := eCtx.Locals(mw.OrganizationIDKey).(string)
	objStorage, err := s.state.GetObjectStorage(orgID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	fileData, err := objStorage.GetObjectData(ctx, bucket, jsonForm.FileName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
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
// @Param X-Organization-Id header string false "Unique Organization ID to choose s3 instance"
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

	span := trace.SpanFromContext(ctx)

	bucket, err := ExtractBucketParameter(eCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	span.SetAttributes(attribute.String("bucket", bucket))

	var jsonForm form.RemoveFileForm
	err = json.Unmarshal(eCtx.Body(), &jsonForm)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	eCtx.Locals("file_path", jsonForm.FileName)

	orgID := eCtx.Locals(mw.OrganizationIDKey).(string)
	objStorage, err := s.state.GetObjectStorage(orgID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	if err = objStorage.DeleteObject(ctx, bucket, jsonForm.FileName); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
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
// @Param X-Organization-Id header string false "Unique Organization ID to choose s3 instance"
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

	span := trace.SpanFromContext(ctx)

	bucket, err := ExtractBucketParameter(eCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	span.SetAttributes(attribute.String("bucket", bucket))

	fileName, err := ExtractFileNameParameter(eCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	span.SetAttributes(attribute.String("file_name", fileName))
	eCtx.Locals("file_path", fileName)

	orgID := eCtx.Locals(mw.OrganizationIDKey).(string)
	objStorage, err := s.state.GetObjectStorage(orgID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	if err = objStorage.DeleteObject(ctx, bucket, fileName); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return eCtx.Status(fiber.StatusOK).SendString("Ok")
}

// CopyFile
// @Summary Copy file to another location into bucket
// @Description Copy file to another location into bucket
// @ID copy-file
// @Tags files
// @Accept  json
// @Produce json
// @Param X-Organization-Id header string false "Unique Organization ID to choose s3 instance"
// @Param bucket path string true "Bucket name of src file"
// @Param jsonQuery body form.CopyFileForm true "Params to copy file"
// @Success 200 {object} form.Success "Ok"
// @Failure	400 {object} form.BadRequestError "Bad Request error"
// @Failure	404 {object} form.NotFoundError "Bucket or file not found"
// @Failure	500 {object} form.InternalServerError "Internal server error"
// @Failure	503 {object} form.ServerUnavailableError "Server does not available"
// @Router /api/v1/cloud/{bucket}/file [patch]
func (s *Server) CopyFile(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()

	span := trace.SpanFromContext(ctx)

	bucket, err := ExtractBucketParameter(eCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	orgID := eCtx.Locals(mw.OrganizationIDKey).(string)
	objStorage, err := s.state.GetObjectStorage(orgID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	exist, err := objStorage.IsBucketExists(ctx, bucket)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(http.StatusBadRequest).SendString(err.Error())
	}

	if !exist {
		err = fmt.Errorf("specified bucket %s does not exist", bucket)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(http.StatusNotFound).SendString(err.Error())
	}

	var jsonForm form.CopyFileForm
	err = json.Unmarshal(eCtx.Body(), &jsonForm)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	eCtx.Locals("file_path", jsonForm.SrcPath)
	params := &domain.CopyObjectParams{
		SourcePath:      jsonForm.SrcPath,
		DestinationPath: jsonForm.DstPath,
	}

	err = objStorage.CopyObject(ctx, bucket, params)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	if jsonForm.WithRemove {
		err = objStorage.DeleteObject(ctx, bucket, params.SourcePath)
		if err != nil {
			err = fmt.Errorf("failed to delete object: %w", err)
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
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
// @Param X-Organization-Id header string false "Unique Organization ID to choose s3 instance"
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

	span := trace.SpanFromContext(ctx)

	bucket, err := ExtractBucketParameter(eCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	span.SetAttributes(attribute.String("bucket", bucket))

	var jsonForm form.GetFilesForm
	err = json.Unmarshal(eCtx.Body(), &jsonForm)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	params := &domain.GetObjectsParams{
		PrefixPath: jsonForm.DirectoryName,
		Limit:      jsonForm.Limit,
		Offset:     jsonForm.Offset,
	}

	orgID := eCtx.Locals(mw.OrganizationIDKey).(string)
	objStorage, err := s.state.GetObjectStorage(orgID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	listObjects, err := objStorage.LoadBucketObjects(ctx, bucket, params)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	// Remove all .keeper meta files from target list objects
	listObjects = slices.DeleteFunc(listObjects, func(object domain.Object) bool {
		return object.Name == FolderFileKeeper
	})

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
// @Param X-Organization-Id header string false "Unique Organization ID to choose s3 instance"
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

	span := trace.SpanFromContext(ctx)

	bucket, err := ExtractBucketParameter(eCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	span.SetAttributes(attribute.String("bucket", bucket))

	var jsonForm form.GetFileAttributesForm
	err = json.Unmarshal(eCtx.Body(), &jsonForm)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	eCtx.Locals("file_path", jsonForm.FilePath)

	orgID := eCtx.Locals(mw.OrganizationIDKey).(string)
	objStorage, err := s.state.GetObjectStorage(orgID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	object, err := objStorage.GetObjectInfo(ctx, bucket, jsonForm.FilePath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
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
// @Param X-Organization-Id header string false "Unique Organization ID to choose s3 instance"
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

	span := trace.SpanFromContext(ctx)

	bucket, err := ExtractBucketParameter(eCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	span.SetAttributes(attribute.String("bucket", bucket))

	var jsonForm form.ShareFileForm
	err = json.Unmarshal(eCtx.Body(), &jsonForm)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	expired := time.Second * time.Duration(jsonForm.ExpiredSecs)
	params := &domain.ShareObjectParams{FilePath: jsonForm.FilePath, Expired: expired}

	eCtx.Locals("file_path", jsonForm.FilePath)

	orgID := eCtx.Locals(mw.OrganizationIDKey).(string)
	objStorage, err := s.state.GetObjectStorage(orgID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	url, err := objStorage.GenShareURL(ctx, bucket, params)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return eCtx.Status(fiber.StatusOK).JSON(form.SuccessResponse(url))
}
