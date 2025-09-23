package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) CreateStorageBucketsGroup() error {
	group := s.server.Group("/api/v1/cloud")

	group.GET("/buckets", s.GetBuckets)
	group.PUT("/bucket", s.CreateBucket)
	group.DELETE("/:bucket", s.RemoveBucket)

	return nil
}

// GetBuckets
// @Summary Get watched bucket list
// @Description Get watched bucket list
// @ID get-buckets
// @Tags buckets
// @Produce  json
// @Success 200 {array} string "Ok"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /cloud/buckets [get]
func (s *Server) GetBuckets(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	watcherDirs, err := s.uc.GetObjectStorage().GetBuckets(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return eCtx.JSON(200, watcherDirs)
}

// CreateBucket
// @Summary Create new bucket into cloud
// @Description Create new bucket into cloud
// @ID create-bucket
// @Tags buckets
// @Accept  json
// @Produce json
// @Param jsonQuery body CreateBucketForm true "Bucket name to create"
// @Success 200 {object} ResponseForm "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /cloud/bucket [put]
func (s *Server) CreateBucket(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()

	jsonForm := &CreateBucketForm{}
	decoder := json.NewDecoder(eCtx.Request().Body)
	err := decoder.Decode(jsonForm)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = s.uc.GetObjectStorage().CreateBucket(ctx, jsonForm.BucketName)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return eCtx.JSON(201, createStatusResponse("Ok"))
}

// RemoveBucket
// @Summary Remove bucket from cloud
// @Description Remove bucket from cloud
// @ID remove-bucket
// @Tags buckets
// @Produce  json
// @Param bucket path string true "Bucket name to remove"
// @Success 200 {object} ResponseForm "Ok"
// @Failure	400 {object} BadRequestForm "Bad Request message"
// @Failure	503 {object} ServerErrorForm "Server does not available"
// @Router /cloud/{bucket} [delete]
func (s *Server) RemoveBucket(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()

	bucket := eCtx.Param("bucket")
	err := s.uc.GetObjectStorage().RemoveBucket(ctx, bucket)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return eCtx.JSON(200, createStatusResponse("Ok"))
}
