package httpserver

import (
	"context"
	"fmt"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/trace"
	"watchtower/internal/application/services/server"
	"watchtower/internal/application/usecase"
	"watchtower/internal/application/utils/telemetry"
	"watchtower/internal/infrastructure/config"
	"watchtower/internal/infrastructure/httpserver/mw"

	echoSwagger "github.com/swaggo/echo-swagger"
	_ "watchtower/docs"
)

type Server struct {
	config *config.ServerConfig
	server *echo.Echo
	tracer trace.Tracer

	processor   *usecase.PipelineUseCase
	storage     *usecase.StorageUseCase
	taskManager *usecase.TaskMangerUseCase
}

func New(
	config *config.ServerConfig,
	tracer trace.Tracer,
	processor *usecase.PipelineUseCase,
	storage *usecase.StorageUseCase,
	taskManager *usecase.TaskMangerUseCase,
) *Server {
	return &Server{
		config: config,
		tracer: tracer,

		processor:   processor,
		storage:     storage,
		taskManager: taskManager,
	}
}

func (s *Server) setupServer() {
	s.server = echo.New()

	s.server.Use(middleware.CORS())
	s.server.Use(middleware.Recover())

	s.initMeterMW()
	s.initLoggerMW()
	s.initTracerMW()

	_ = s.CreateTasksGroup()
	_ = s.CreateStorageBucketsGroup()
	_ = s.CreateStorageObjectsGroup()

	s.server.GET("/api/v1/swagger/*", echoSwagger.WrapHandler)
}

func (s *Server) Start(_ context.Context) error {
	s.setupServer()
	if err := s.server.Start(s.config.Http.Address); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) initMeterMW() {
	s.server.Use(echoprometheus.NewMiddleware(server.AppName))
	s.server.GET("/metrics", echoprometheus.NewHandler())
}

func (s *Server) initLoggerMW() {
	if s.config.Logger.EnableLoki {
		lokiLog := telemetry.InitLokiLogger(s.config.Logger)
		s.server.Use(mw.CreateLokiLoggerMW(&lokiLog))
	} else {
		s.server.Use(mw.InitLocalLogger(s.config.Logger))
	}
}

func (s *Server) initTracerMW() {
	s.server.Use(otelecho.Middleware(
		server.AppName,
		otelecho.WithPropagators(telemetry.TracePropagator),
		otelecho.WithSkipper(mw.TracerSkipper),
	))
}
