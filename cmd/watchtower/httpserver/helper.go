package httpserver

import (
	"fmt"
	"log/slog"
	"mime/multipart"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func ExtractBucketParameter(eCtx *fiber.Ctx) (string, error) {
	bucket := eCtx.Params("bucket")
	if bucket == "" {
		err := fmt.Errorf("bucket parameter is required")
		return "", err
	}

	return bucket, nil
}

func ExtractTaskIDParameter(eCtx *fiber.Ctx) (uuid.UUID, error) {
	taskIDParam := eCtx.Params("task_id")
	if taskIDParam == "" {
		err := fmt.Errorf("bucket parameter is required")
		return uuid.Nil, err
	}

	taskID, err := uuid.Parse(taskIDParam)
	if err != nil {
		return taskID, err
	}

	return taskID, nil
}

func ExtractTaskStatusParameter(eCtx *fiber.Ctx) (int, error) {
	statusParam := eCtx.Query("status")
	status, err := strconv.Atoi(statusParam)
	if err != nil {
		err = fmt.Errorf("unknown status parameter: %w", err)
		return -1, err
	}

	return status, nil
}

func ExtractFileNameParameter(eCtx *fiber.Ctx) (string, error) {
	fileNameQuery := eCtx.Query("file_name")
	if fileNameQuery == "" {
		err := fmt.Errorf("bucket parameter is required")
		return "", err
	}

	return fileNameQuery, nil
}

func ExtractFilePrefixParameter(eCtx *fiber.Ctx) string {
	return eCtx.FormValue("prefix", "./")
}

func ExtractMultipartForm(eCtx *fiber.Ctx) (*multipart.Form, error) {
	multipartForm, err := eCtx.MultipartForm()
	if err != nil {
		return nil, err
	}

	if multipartForm.File["files"] == nil {
		err = fmt.Errorf("there are no files into multipart form")
		return nil, err
	}

	return multipartForm, nil
}

func ExtractExpiredDatetime(eCtx *fiber.Ctx) (*time.Time, error) {
	expired := eCtx.Query("expired")
	if expired == "" {
		slog.Debug("expired parameter has not been set")
		return nil, nil
	}

	timeVal, err := time.Parse(time.RFC3339, expired)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expired datetime: %w", err)
	}

	return &timeVal, nil
}
