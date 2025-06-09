package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogloki "github.com/samber/slog-loki/v2"
)

const AppName = "watchtower"

type SlogLokiLogger struct {
	client    *slog.Logger
	filterURI []string
}

func InitLocalLogger(config LoggerConfig) echo.MiddlewareFunc {
	logConfig := middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			uri := c.Path()
			return strings.Contains(uri, "swagger")
		},

		Format: fmt.Sprintf(
			"%s  %s %s request{%s}: %s %s ms %s\n",
			"${time_rfc3339}",
			config.Level,
			"${id}",
			"method=${method} uri=${path}",
			"latency=${latency}",
			"status=${status}",
			"error=\"${error}\"",
		),
	}

	return middleware.LoggerWithConfig(logConfig)
}

func InitLokiLogger(config LoggerConfig) SlogLokiLogger {
	lokiConfig := slogloki.Option{
		Endpoint:           fmt.Sprintf("%s/api/prom/push", config.Address),
		Level:              slog.LevelInfo,
		BatchWait:          time.Second * 5,
		BatchEntriesNumber: 10,
	}

	logger := slog.New(lokiConfig.NewLokiHandler()).
		With("service_name", AppName).
		With("service", AppName).
		With("detected_level", config.Level).
		With("level", config.Level)

	filterURI := []string{
		"/metrics",
		"/swagger/*",
	}

	return SlogLokiLogger{client: logger, filterURI: filterURI}
}

func (sll *SlogLokiLogger) LokiLoggerMW() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if slices.Contains(sll.filterURI, c.Path()) {
				return next(c)
			}

			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			latency := time.Since(start)

			logMessage := map[string]interface{}{
				"message":    c.Response().Status,
				"latency":    latency.String(),
				"status":     c.Response().Status,
				"method":     c.Request().Method,
				"uri":        c.Path(),
				"client_ip":  c.RealIP(),
				"user_agent": c.Request().UserAgent(),
			}
			jsonMessage, _ := json.Marshal(logMessage)

			var logLevel slog.Level
			statusCategory := c.Response().Status / 100
			if statusCategory < 3 && statusCategory >= 2 {
				logLevel = slog.LevelInfo
			} else {
				logLevel = slog.LevelError
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			sll.client.Log(ctx, logLevel, string(jsonMessage))
			defer cancel()

			return err
		}
	}
}
