package httpserver

import (
	"context"
	"fmt"
	v1 "watchtower/internal/infrastructure/httpserver/api/v1"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"watchtower/internal/application/usecase"
)

type Server struct {
	e      *echo.Echo
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
	s.e = echo.New()

	s.e.Use(echoprometheus.NewMiddleware(AppName))
	s.e.GET("/metrics", echoprometheus.NewHandler())

	if s.config.Logger.EnableLoki {
		lokiLog := InitLokiLogger(s.config.Logger)
		s.e.Use(lokiLog.LokiLoggerMW())
	} else {
		s.e.Use(InitLocalLogger(s.config.Logger))
	}

	s.e.Use(middleware.CORS())
	s.e.Use(middleware.Recover())

	v1.SetupV1(s.e, s.uc)
}

func (s *Server) Start(_ context.Context) error {
	s.setupServer()
	if err := s.e.Start(s.config.Address); err != nil {
		return fmt.Errorf("failed to start e: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}
