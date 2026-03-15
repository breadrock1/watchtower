package httpserver

import (
	"fmt"

	"go.opentelemetry.io/otel/trace"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"

	"watchtower/cmd/watchtower/httpserver/mw"
	"watchtower/internal/process"
	"watchtower/internal/shared/kernel"
	"watchtower/internal/shared/telemetry"

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
	Server *fiber.App
}

func SetupServer(
	otlpConfig telemetry.OtlpConfig,
	state *process.Orchestrator,
	tracer trace.Tracer,
) *Server {
	serverApp := &Server{
		tracer: tracer,
		state:  state,
	}

	serverApp.Server = fiber.New()

	serverApp.Server.Use(cors.New(cors.Config{}))
	serverApp.Server.Use(recover.New())

	serverApp.initMeterMW()
	serverApp.initTracerMW(otlpConfig.Tracer)
	serverApp.initLoggerMW(otlpConfig.Logger)

	serverApp.Server.Get("/monitor", monitor.New())

	api := serverApp.Server.Group("/api")
	api.Get("/swagger/*", swagger.HandlerDefault)

	v1Api := api.Group("/v1")

	serverApp.CreateSystemGroup(v1Api)
	serverApp.CreateTasksGroup(v1Api)
	serverApp.CreateStorageBucketsGroup(v1Api)
	serverApp.CreateStorageObjectsGroup(v1Api)

	return serverApp
}

func (s *Server) Start(_ kernel.Ctx, config Config) error {
	if err := s.Server.Listen(config.Address); err != nil {
		return fmt.Errorf("failed to start Server: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(_ kernel.Ctx) error {
	return s.Server.Shutdown()
}

func (s *Server) initMeterMW() {
	prometheus := fiberprometheus.New(telemetry.AppName)
	prometheus.RegisterAt(s.Server, "/metrics")
	prometheus.SetSkipPaths([]string{"/swagger"})
	prometheus.SetIgnoreStatusCodes([]int{401, 403, 404})
	s.Server.Use(prometheus.Middleware)
}

func (s *Server) initLoggerMW(logConfig telemetry.LoggerConfig) {
	if logConfig.EnableLoki {
		lokiLog := telemetry.InitLokiLogger(logConfig)
		s.Server.Use(mw.CreateLokiLoggerMW(&lokiLog))
	} else {
		s.Server.Use(mw.InitLocalLogger(logConfig))
	}
}

func (s *Server) initTracerMW(_ telemetry.TracerConfig) {
	s.Server.Use(otelfiber.Middleware(
		otelfiber.WithNext(mw.TracerURLSkipper),
	))
}
