package mw

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"watchtower/internal/shared/telemetry"
)

func InitLocalLogger(config telemetry.LoggerConfig) fiber.Handler {
	return logger.New(
		logger.Config{
			TimeFormat: "2006/01/02 15:04:05",
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
			Next: func(eCtx *fiber.Ctx) bool {
				uri := eCtx.Request().URI().String()
				excludeSwagger := strings.Contains(uri, "swagger")
				excludeMetrics := strings.Contains(uri, "metrics")
				return excludeSwagger || excludeMetrics
			},
		},
	)
}

func CreateLokiLoggerMW(sll *telemetry.SlogLokiLogger) fiber.Handler {
	return func(eCtx *fiber.Ctx) error {
		if slices.Contains(sll.FilterURI, eCtx.Path()) {
			return eCtx.Next()
		}

		start := time.Now()

		err := eCtx.Next()
		if err != nil {
			return err
		}

		latency := time.Since(start)

		var responseMsg string
		if eCtx.Response().StatusCode() >= 200 {
			responseMsg = "Ok"
		} else {
			responseMsg = eCtx.Response().String()
		}

		logMessage := map[string]interface{}{
			"message":    responseMsg,
			"latency":    latency.String(),
			"status":     eCtx.Response().StatusCode(),
			"method":     eCtx.Method(),
			"uri":        eCtx.Path(),
			"client_ip":  eCtx.IP(),
			"user_agent": eCtx.Request(),
		}
		jsonMessage, _ := json.Marshal(logMessage)

		var logLevel slog.Level
		statusCategory := eCtx.Response().StatusCode() / 100
		if statusCategory < 3 && statusCategory >= 2 {
			logLevel = slog.LevelInfo
		} else {
			logLevel = slog.LevelError
		}

		ctx, cancel := context.WithTimeout(eCtx.Context(), 5*time.Second)
		sll.Client.Log(ctx, logLevel, string(jsonMessage))
		defer cancel()

		return err
	}
}
