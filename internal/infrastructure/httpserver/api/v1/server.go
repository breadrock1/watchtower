package v1

import (
	"github.com/labstack/echo/v4"
	"watchtower/internal/application/usecase"

	echoSwagger "github.com/swaggo/echo-swagger"
	_ "watchtower/docs"
)

type V1Server struct {
	e  *echo.Echo
	uc *usecase.UseCase
}

func SetupV1(server *echo.Echo, uc *usecase.UseCase) {
	v1Server := &V1Server{
		e:  server,
		uc: uc,
	}

	_ = v1Server.CreateTasksGroup()
	_ = v1Server.CreateStorageBucketsGroup()
	_ = v1Server.CreateStorageObjectsGroup()

	v1Server.e.GET("/api/v1/swagger/*", echoSwagger.WrapHandler)
}
