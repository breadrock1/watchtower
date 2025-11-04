package httpserver

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/trace"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"watchtower/cmd/watchtower/httpserver/mw"
	"watchtower/internal/process"
	"watchtower/internal/shared/telemetry"

	_ "watchtower/docs"

	echoSwagger "github.com/swaggo/echo-swagger"
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

	state  *process.Orchestrator
	server *echo.Echo
}

func SetupServer(
	config telemetry.OtlpConfig,
	state *process.Orchestrator,
	tracer trace.Tracer,
) *Server {
	serverApp := &Server{
		tracer: tracer,
		state:  state,
	}

	serverApp.server = echo.New()

	serverApp.server.Use(middleware.CORS())
	serverApp.server.Use(middleware.Recover())

	serverApp.initMeterMW()
	serverApp.initTracerMW()
	serverApp.initLoggerMW(config.Logger)

	_ = serverApp.CreateTasksGroup()
	_ = serverApp.CreateStorageBucketsGroup()
	_ = serverApp.CreateStorageObjectsGroup()

	serverApp.server.GET("/api/v1/swagger/*", echoSwagger.WrapHandler)

	return serverApp
}

func (s *Server) Start(_ context.Context, config Config) error {
	if err := s.server.Start(config.Address); err != nil {
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

func (s *Server) initLoggerMW(logConfig telemetry.LoggerConfig) {
	if logConfig.EnableLoki {
		lokiLog := telemetry.InitLokiLogger(logConfig)
		s.server.Use(mw.CreateLokiLoggerMW(&lokiLog))
	} else {
		s.server.Use(mw.InitLocalLogger(logConfig))
	}
}

func (s *Server) initTracerMW() {
	s.server.Use(otelecho.Middleware(
		telemetry.AppName,
		otelecho.WithPropagators(telemetry.TracePropagator),
		otelecho.WithSkipper(mw.TracerSkipper),
	))
}
