package mw

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
	"watchtower/internal/application/utils/telemetry"
)

func InitLocalLogger(config telemetry.LoggerConfig) echo.MiddlewareFunc {
	logConfig := middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			uri := c.Path()
			return strings.Contains(uri, "swagger")
		},
		CustomTimeFormat: "2006/01/02 15:04:05",
		Format: fmt.Sprintf(
			"%s %s http-response={%s %s %s %s %s}\n",
			"${time_custom}",
			config.Level,
			"method=${method}",
			"uri=${path}",
			"latency=${latency}",
			"status=${status}",
			"error=\"${error}\"",
		),
	}

	return middleware.LoggerWithConfig(logConfig)
}

func CreateLokiLoggerMW(sll *telemetry.SlogLokiLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(eCtx echo.Context) error {
			if slices.Contains(sll.FilterURI, eCtx.Path()) {
				return next(eCtx)
			}

			start := time.Now()

			err := next(eCtx)
			if err != nil {
				eCtx.Error(err)
			}

			latency := time.Since(start)

			logMessage := map[string]interface{}{
				"message":    eCtx.Response().Status,
				"latency":    latency.String(),
				"status":     eCtx.Response().Status,
				"method":     eCtx.Request().Method,
				"uri":        eCtx.Path(),
				"client_ip":  eCtx.RealIP(),
				"user_agent": eCtx.Request().UserAgent(),
			}
			jsonMessage, _ := json.Marshal(logMessage)

			var logLevel slog.Level
			statusCategory := eCtx.Response().Status / 100
			if statusCategory < 3 && statusCategory >= 2 {
				logLevel = slog.LevelInfo
			} else {
				logLevel = slog.LevelError
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			sll.Client.Log(ctx, logLevel, string(jsonMessage))
			defer cancel()

			return err
		}
	}
}
