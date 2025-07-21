package httpserver

import (
	"context"
	"fmt"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"watchtower/internal/application/usecase"

	echoSwagger "github.com/swaggo/echo-swagger"
	_ "watchtower/docs"
)

type Server struct {
	server *echo.Echo
	config *Config

	uc *usecase.UseCase
}

func New(config *Config, watcherUC *usecase.UseCase) *Server {
	return &Server{
		config: config,
		uc:     watcherUC,
	}
}

func (s *Server) setupServer() {
	s.server = echo.New()

	s.server.Use(echoprometheus.NewMiddleware(AppName))
	s.server.GET("/metrics", echoprometheus.NewHandler())

	if s.config.Logger.EnableLoki {
		lokiLog := InitLokiLogger(s.config.Logger)
		s.server.Use(lokiLog.LokiLoggerMW())
	}
	s.server.Use(InitLocalLogger(s.config.Logger))

	s.server.Use(middleware.CORS())
	s.server.Use(middleware.Recover())

	_ = s.CreateTasksGroup()
	_ = s.CreateStorageBucketsGroup()
	_ = s.CreateStorageObjectsGroup()

	s.server.GET("/swagger/*", echoSwagger.WrapHandler)
}

func (s *Server) Start(_ context.Context) error {
	s.setupServer()
	if err := s.server.Start(s.config.Address); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
