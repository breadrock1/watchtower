package telemetry

import (
	"fmt"
	"log/slog"
	"time"

	slogloki "github.com/samber/slog-loki/v2"
)

var (
	filterURI = []string{
		"/metrics",
		"/swagger/*",
	}
)

type SlogLokiLogger struct {
	Client    *slog.Logger
	FilterURI []string
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

	return SlogLokiLogger{Client: logger, FilterURI: filterURI}
}
