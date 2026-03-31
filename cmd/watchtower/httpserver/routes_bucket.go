package httpserver

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"watchtower/cmd/watchtower/httpserver/form"
)

func (s *Server) CreateStorageBucketsGroup(group fiber.Router) {
	group.Get("/cloud/buckets", s.GetBuckets)
	group.Put("/cloud/bucket", s.CreateBucket)
	group.Delete("/cloud/:bucket", s.RemoveBucket)
}

// GetBuckets
// @Summary Get watched bucket list
// @Description Get watched bucket list
// @ID get-buckets
// @Tags buckets
// @Produce  json
// @Success 200 {object} []form.BucketSchema "Loaded buckets info"
// @Failure	500 {object} form.InternalServerError "Internal server error"
// @Failure	503 {object} form.ServerUnavailableError "Server does not available"
// @Router /api/v1/cloud/buckets [get]
func (s *Server) GetBuckets(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()

	span := trace.SpanFromContext(ctx)

	objStorage := s.state.GetObjectStorage()
	buckets, err := objStorage.GetAllBuckets(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	bucketsDto := make([]form.BucketSchema, len(buckets))
	for index, bucket := range buckets {
		bucketsDto[index] = form.BucketFromDomain(bucket)
	}

	return eCtx.Status(fiber.StatusOK).JSON(bucketsDto)
}

// CreateBucket
// @Summary Create new bucket into cloud
// @Description Create new bucket into cloud
// @ID create-bucket
// @Tags buckets
// @Accept  json
// @Produce json
// @Param jsonQuery body form.CreateBucketForm true "Bucket name to create"
// @Success 200 {object} form.Success "Ok"
// @Failure	400 {object} form.BadRequestError "Bad Request error"
// @Failure	500 {object} form.InternalServerError "Internal server error"
// @Failure	503 {object} form.ServerUnavailableError "Server does not available"
// @Router /api/v1/cloud/bucket [put]
func (s *Server) CreateBucket(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()

	span := trace.SpanFromContext(ctx)

	var jsonForm form.CreateBucketForm
	err := json.Unmarshal(eCtx.Body(), &jsonForm)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	objStorage := s.state.GetObjectStorage()
	exists, err := objStorage.IsBucketExists(ctx, jsonForm.BucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	if exists {
		span.SetStatus(codes.Error, "bucket already exists")
		span.RecordError(err)
		// TODO: Temporary solution. Need to return 409 http error
		// return eCtx.Status(fiber.StatusConflict).SendString("bucket already exists")
		return eCtx.Status(fiber.StatusOK).SendString("bucket already exists")
	}

	err = s.state.GetObjectStorage().CreateBucket(ctx, jsonForm.BucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return eCtx.Status(fiber.StatusCreated).SendString("Ok")
}

// RemoveBucket
// @Summary Remove bucket from cloud
// @Description Remove bucket from cloud
// @ID remove-bucket
// @Tags buckets
// @Produce  json
// @Param bucket path string true "Bucket name to remove"
// @Success 200 {object} form.Success "Ok"
// @Failure	400 {object} form.BadRequestError "Bad Request error"
// @Failure	404 {object} form.NotFoundError "Bucket not found"
// @Failure	500 {object} form.InternalServerError "Internal server error"
// @Failure	503 {object} form.ServerUnavailableError "Server does not available"
// @Router /api/v1/cloud/{bucket} [delete]
func (s *Server) RemoveBucket(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()

	span := trace.SpanFromContext(ctx)

	bucket, err := ExtractBucketParameter(eCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	span.SetAttributes(attribute.String("bucket", bucket))

	objStorage := s.state.GetObjectStorage()
	exists, err := objStorage.IsBucketExists(ctx, bucket)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	if !exists {
		span.SetStatus(codes.Error, "buket does not exist")
		span.RecordError(err)
		// TODO: Temporary solution. Need to return 409 http error
		// return eCtx.Status(fiber.StatusConflict).SendString("bucket already exists")
		return eCtx.Status(fiber.StatusNotFound).SendString("bucket already exists")
	}

	err = objStorage.DeleteBucket(ctx, bucket)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return eCtx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return eCtx.Status(fiber.StatusOK).SendString("Ok")
}
