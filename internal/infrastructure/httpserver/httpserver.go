package httpserver

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/trace"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"watchtower/internal/application/service/server"
	"watchtower/internal/application/utils/telemetry"
	"watchtower/internal/infrastructure/config"
	"watchtower/internal/infrastructure/httpserver/mw"

	echoSwagger "github.com/swaggo/echo-swagger"
	_ "watchtower/docs"
)

// Server
// @title          Watchtower service
// @version        0.0.1
// @description    Watchtower is a project designed to provide processing files created into cloud by events.
//
// @tag.name tasks
// @tag.description APIs to get status tasks. When TaskStatus may be:
//
//	Failed -> -1;
//	Received -> 0;
//	Pending -> 1;
//	processConsumedTask -> 2;
//	Successful -> 3.
//
// @host      localhost:2893
// @BasePath  /api/v1
//
// @tag.name buckets
// @tag.description CRUD APIs to manage cloud buckets
//
// @tag.name files
// @tag.description CRUD APIs to manage files into bucket
//
// @tag.name share
// @tag.description Share files by URL API
type Server struct {
	tracer trace.Tracer

	config *config.ServerConfig
	state  *server.ServerState
	server *echo.Echo
}

func New(
	config *config.ServerConfig,
	state *server.ServerState,
	tracer trace.Tracer,
) *Server {
	return &Server{
		config: config,
		tracer: tracer,
		state:  state,
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
	s.server.Use(echoprometheus.NewMiddleware(telemetry.AppName))
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
		telemetry.AppName,
		otelecho.WithPropagators(telemetry.TracePropagator),
		otelecho.WithSkipper(mw.TracerSkipper),
	))
}
