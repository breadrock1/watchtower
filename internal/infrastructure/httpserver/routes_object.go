package httpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"watchtower/internal/application/dto"
)

func (s *Server) CreateStorageObjectsGroup() error {
	group := s.server.Group("/cloud")

	group.POST("/:bucket/files", s.GetFiles)
	group.POST("/:bucket/file/copy", s.CopyFile)
	group.POST("/:bucket/file/move", s.MoveFile)
	group.PUT("/:bucket/file/upload", s.UploadFile)
	group.POST("/:bucket/file/download", s.DownloadFile)
	group.DELETE("/:bucket/file", s.RemoveFile2)
	group.DELETE("/:bucket/file/remove", s.RemoveFile)
	group.POST("/:bucket/file/attributes", s.GetFileAttributes)

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
// @Param jsonQuery body CopyFileForm true "Params to copy file"
// @Success 200 {object} ResponseForm "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /cloud/{bucket}/file/copy [post]
func (s *Server) CopyFile(eCtx echo.Context) error {
	bucket := eCtx.Param("bucket")

	jsonForm := &CopyFileForm{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	err := decoder.Decode(jsonForm)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := eCtx.Request().Context()
	err = s.uc.GetObjectStorage().CopyFile(ctx, bucket, jsonForm.SrcPath, jsonForm.DstPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return eCtx.JSON(200, createStatusResponse("Ok"))
}

// MoveFile
// @Summary Move file to another location into bucket
// @Description Move file to another location into bucket
// @ID move-file
// @Tags files
// @Accept  json
// @Produce json
// @Param bucket path string true "Bucket name of src file"
// @Param jsonQuery body CopyFileForm true "Params to move file"
// @Success 200 {object} ResponseForm "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /cloud/{bucket}/file/move [post]
func (s *Server) MoveFile(eCtx echo.Context) error {
	bucket := eCtx.Param("bucket")

	jsonForm := &CopyFileForm{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	err := decoder.Decode(jsonForm)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := eCtx.Request().Context()
	err = s.uc.GetObjectStorage().CopyFile(ctx, bucket, jsonForm.SrcPath, jsonForm.DstPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return eCtx.JSON(200, createStatusResponse("Ok"))
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
// @Success 200 {object} ResponseForm "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /cloud/{bucket}/file/upload [put]
func (s *Server) UploadFile(eCtx echo.Context) error {
	var fileData bytes.Buffer

	multipartForm, err := eCtx.MultipartForm()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	bucket := eCtx.Param("bucket")
	if exist, err := s.uc.GetObjectStorage().IsBucketExist(eCtx.Request().Context(), bucket); err != nil || !exist {
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
		log.Println("failed to parse expired time param: ", expired, timeParseErr)
	}

	uploadedFiles := make([]*dto.TaskEvent, len(multipartForm.File["files"]))
	ctx := eCtx.Request().Context()
	for index, fileForm := range multipartForm.File["files"] {
		fileName := fileForm.Filename
		fileHandler, err := fileForm.Open()
		if err != nil {
			log.Println("failed to open file form", err)
			continue
		}
		defer func() {
			if err := fileHandler.Close(); err != nil {
				log.Println("failed to close file handler: ", fileName, err)
				return
			}
		}()

		fileData.Reset()
		_, err = fileData.ReadFrom(fileHandler)
		if err != nil {
			log.Println("failed to read file form", fileName, err)
			continue
		}

		uploadItem := dto.FileToUpload{
			Bucket:   bucket,
			FilePath: fileName,
			FileData: &fileData,
			Expired:  &timeVal,
		}

		task, err := s.uc.StoreFileToStorage(ctx, uploadItem)
		if err != nil {
			log.Println("failed to upload file to cloud: ", fileName, err)
			continue
		}

		uploadedFiles[index] = task
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
// @Param jsonQuery body DownloadFileForm true "Parameters to download file"
// @Success 200 {file} io.Writer "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /cloud/{bucket}/file/download [post]
func (s *Server) DownloadFile(eCtx echo.Context) error {
	bucket := eCtx.Param("bucket")

	jsonForm := &DownloadFileForm{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	if err := decoder.Decode(jsonForm); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := eCtx.Request().Context()
	fileData, err := s.uc.GetObjectStorage().DownloadFile(ctx, bucket, jsonForm.FileName)
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
// @Param jsonQuery body RemoveFileForm true "Parameters to remove file"
// @Success 200 {object} ResponseForm "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /cloud/{bucket}/file/remove [delete]
func (s *Server) RemoveFile(eCtx echo.Context) error {
	bucket := eCtx.Param("bucket")

	jsonForm := &RemoveFileForm{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	if err := decoder.Decode(jsonForm); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := eCtx.Request().Context()
	if err := s.uc.GetObjectStorage().DeleteFile(ctx, bucket, jsonForm.FileName); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return eCtx.JSON(200, createStatusResponse("Ok"))
}

// RemoveFile2
// @Summary Remove file from cloud
// @Description Remove file from cloud
// @ID remove-file-2
// @Tags files
// @Produce  json
// @Param bucket path string true "Bucket name to remove file"
// @Param file_name query string true "Parameters to remove file"
// @Success 200 {object} ResponseForm "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /cloud/{bucket}/file [delete]
func (s *Server) RemoveFile2(eCtx echo.Context) error {
	bucket := eCtx.Param("bucket")
	fileName := eCtx.QueryParam("file_name")
	ctx := eCtx.Request().Context()
	if err := s.uc.GetObjectStorage().DeleteFile(ctx, bucket, fileName); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return eCtx.JSON(200, createStatusResponse("Ok"))
}

// GetFiles
// @Summary Get files list into bucket
// @Description Get files list into bucket
// @ID get-list-files
// @Tags files
// @Accept  json
// @Produce json
// @Param bucket path string true "Bucket name to get list files"
// @Param jsonQuery body GetFilesForm true "Parameters to get list files"
// @Success 200 {object} ResponseForm "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /cloud/{bucket}/files [post]
func (s *Server) GetFiles(eCtx echo.Context) error {
	bucket := eCtx.Param("bucket")

	jsonForm := &GetFilesForm{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	err := decoder.Decode(jsonForm)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := eCtx.Request().Context()
	listObjects, err := s.uc.GetObjectStorage().GetBucketFiles(ctx, bucket, jsonForm.DirectoryName)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return eCtx.JSON(200, listObjects)
}

// GetFileAttributes
// @Summary Get file attributes
// @Description Get file attributes
// @ID get-file-attrs
// @Tags files
// @Accept  json
// @Produce json
// @Param bucket path string true "Bucket name to get list files"
// @Param jsonQuery body GetFileAttributesForm true "Parameters to get list files"
// @Success 200 {object} ResponseForm "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /cloud/{bucket}/file/attributes [post]
func (s *Server) GetFileAttributes(eCtx echo.Context) error {
	bucket := eCtx.Param("bucket")

	jsonForm := &GetFileAttributesForm{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	err := decoder.Decode(jsonForm)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := eCtx.Request().Context()
	listObjects, err := s.uc.GetObjectStorage().GetFileMetadata(ctx, bucket, jsonForm.FilePath)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return eCtx.JSON(200, listObjects)
}

// ShareFile
// @Summary Get share URL for file
// @Description Get share URL for file
// @ID share-file
// @Tags share
// @Accept  json
// @Produce json
// @Param bucket path string true "Bucket name to share file"
// @Param jsonQuery body ShareFileForm true "Parameters to share file"
// @Success 200 {object} ResponseForm "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /cloud/{bucket}/file/share [post]
func (s *Server) ShareFile(eCtx echo.Context) error {
	bucket := eCtx.Param("bucket")

	jsonForm := &ShareFileForm{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	err := decoder.Decode(jsonForm)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	expired := time.Second * time.Duration(jsonForm.ExpiredSecs)

	ctx := eCtx.Request().Context()
	url, err := s.uc.GetObjectStorage().GenSharedURL(ctx, expired, bucket, jsonForm.FileName, jsonForm.RedirectHost)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return eCtx.JSON(200, createStatusResponse(url))
}
