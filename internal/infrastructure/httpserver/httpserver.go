package httpserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/otel/trace"
	"watchtower/internal/application/usecase"

	echoSwagger "github.com/swaggo/echo-swagger"
	_ "watchtower/docs"
)

type Server struct {
	server *echo.Echo
	tracer trace.Tracer
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
	} else {
		s.server.Use(InitLocalLogger(s.config.Logger))
	}

	if s.config.Tracer.EnableJaeger {
		tp, err := InitTracer(s.config)
		if err != nil {
			slog.Error("failed to initialize tracer", slog.String("err", err.Error()))
		} else {
			s.tracer = tp
			defer func() {
				//if err := s.tracer.Cleanup(context.Background()); err != nil {
				//	slog.Error("error shutting down tracer provider", slog.String("err", err.Error()))
				//}
				//s.tracer.Shutdown(context.Background())
			}()

			//traceFilterMW := otelecho.Middleware(
			//	AppName,
			//	otelecho.WithTracerProvider(otel.GetTracerProvider()),
			//	otelecho.WithPropagators(otel.GetTextMapPropagator()),
			//	otelecho.WithSkipper(func(c echo.Context) bool {
			//		return shouldSkipTrace(c.Path())
			//	}),
			//)
			//
			//s.server.Use(tracingFilter())
		}
	}

	_, span := s.tracer.Start(context.Background(), "init-watchtower")
	span.End()

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
