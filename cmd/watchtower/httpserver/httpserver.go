package httpserver

import (
	"fmt"
	"log/slog"

	"github.com/breadrock1/otlp-go/otlp"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/trace"

	_ "watchtower/docs"
	"watchtower/internal/process"
	"watchtower/internal/shared/kernel"

	otlppfiber "github.com/breadrock1/otlp-go/pkg/fiber"
)

const (
	// HeaderBufferSize Increase this value to accommodate larger headers
	// e.g., 8192 (8KB), 16384 (16KB), or 32768 (32KB)
	HeaderBufferSize = 8192

	// MultipartBodyLimit Increase this value to accommodate larger request body
	// e.g., 52428800 (50 Mb) 104857600 (100 Mb)
	MultipartBodyLimit = 104857600
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

func SetupServer(otlpConfig otlp_go.OtlpConfig, state *process.Orchestrator) *Server {
	tracer, err := otlp_go.InitTracer(otlpConfig.Tracer)
	if err != nil {
		slog.Warn("failed to init tracer", slog.String("err", err.Error()))
	}

	serverApp := &Server{
		tracer: tracer,
		state:  state,
	}

	serverApp.Server = fiber.New(
		fiber.Config{
			DisableStartupMessage: true,
			ReadBufferSize:        HeaderBufferSize,
			BodyLimit:             MultipartBodyLimit,
		},
	)

	serverApp.initMiddlewares(otlpConfig)

	serverApp.Server.Get("/", serverApp.Home)
	serverApp.Server.Get("/monitor", monitor.New())
	serverApp.Server.Get("/processing/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	api := serverApp.Server.Group("/api")

	api.Get("/swagger/*", swagger.HandlerDefault)

	v1Api := api.Group("/v1")
	serverApp.CreateSystemGroup(v1Api)
	serverApp.CreateTasksGroup(v1Api)
	serverApp.CreateStorageBucketsGroup(v1Api)
	serverApp.CreateStorageObjectsGroup(v1Api)

	return serverApp
}

func (s *Server) Start(config Config) error {
	slog.Info("starting http server", slog.String("address", config.Address))
	if err := s.Server.Listen(config.Address); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx kernel.Ctx) error {
	slog.Info("http server shutting down")
	return s.Server.ShutdownWithContext(ctx)
}

func (s *Server) initMiddlewares(otlpConfig otlp_go.OtlpConfig) {
	s.Server.Use(cors.New(cors.Config{}))
	s.Server.Use(recover.New())

	s.Server.Use(otlppfiber.PrometheusMeterMiddleware(s.Server))
	s.Server.Use(otlppfiber.OtlpJaegerTracerMiddleware())

	logger := otlp_go.InitLocalLogger(otlpConfig.Logger)
	slog.SetDefault(logger)

	s.Server.Use(otlppfiber.StdoutLoggerMiddleware(otlpConfig.Logger))
	if otlpConfig.Logger.EnableLoki {
		s.Server.Use(otlppfiber.RemoteLokiLoggerMiddleware(otlpConfig.Logger))
	}
}
