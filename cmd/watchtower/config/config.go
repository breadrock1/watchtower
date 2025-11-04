package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"watchtower/cmd/watchtower/httpserver"
	"watchtower/internal/core/cloud/infrastructure/s3"
	"watchtower/internal/process"
	"watchtower/internal/shared/telemetry"
	"watchtower/internal/support/task/infrastructure/docparser"
	"watchtower/internal/support/task/infrastructure/docsearch"
	"watchtower/internal/support/task/infrastructure/redis"
	"watchtower/internal/support/task/infrastructure/rmq"
)

type Config struct {
	Orchestrator process.Config       `mapstructure:"orchestrator"`
	Otlp         telemetry.OtlpConfig `mapstructure:"otlp"`
	Server       ServerConfig         `mapstructure:"server"`
	Storage      StorageConfig        `mapstructure:"storage"`
	Task         TaskConfig           `mapstructure:"task"`
}

type ServerConfig struct {
	Http httpserver.Config `mapstructure:"http"`
}

type StorageConfig struct {
	S3 s3.Config `mapstructure:"s3"`
}

type TaskConfig struct {
	TaskStorage TaskStorageConfig `mapstructure:"storage"`
	TaskQueue   TaskQueueConfig   `mapstructure:"queue"`
	Processor   ProcessorConfig   `mapstructure:"processor"`
}

type TaskStorageConfig struct {
	Redis redis.Config `mapstructure:"redis"`
}

type TaskQueueConfig struct {
	Rmq rmq.Config `mapstructure:"rmq"`
}

type ProcessorConfig struct {
	DocParser  docparser.Config `mapstructure:"docparser"`
	DocStorage docsearch.Config `mapstructure:"docstorage"`
}

func FromFile(filePath string) (*Config, error) {
	_ = godotenv.Load()

	config := &Config{}

	viperInstance := viper.New()
	viperInstance.SetConfigFile(filePath)
	viperInstance.SetConfigType("toml")

	viperInstance.AutomaticEnv()
	viperInstance.SetEnvPrefix("watchtower")
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))

	// Orchestrator config
	bindErr := viperInstance.BindEnv("orchestrator.semaphore_size", "WATCHTOWER__ORCHESTRATOR__SEMAPHORE_SIZE")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}

	// Otlp config
	bindErr = viperInstance.BindEnv("otlp.logger.level", "WATCHTOWER__OTLP__LOGGER__LEVEL")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("otlp.logger.address", "WATCHTOWER__OTLP__LOGGER__ADDRESS")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("otlp.logger.enable_loki", "WATCHTOWER__OTLP__LOGGER__ENABLE_LOKI")

	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("otlp.tracer.address", "WATCHTOWER__OTLP__TRACER__ADDRESS")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("otlp.tracer.enable_jaeger", "WATCHTOWER__OTLP__TRACER__ENABLE_JAEGER")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}

	// Server config
	bindErr = viperInstance.BindEnv("server.http.address", "WATCHTOWER__SERVER__HTTP__ADDRESS")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}

	// Storage s3 config
	bindErr = viperInstance.BindEnv("storage.s3.address", "WATCHTOWER__STORAGE__S3__ADDRESS")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("storage.s3.access_id", "WATCHTOWER__STORAGE__S3__ACCESS_ID")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("storage.s3.secret_key", "WATCHTOWER__STORAGE__S3__SECRET_KEY")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("storage.s3.enable_ssl", "WATCHTOWER__STORAGE__S3__ENABLE_SSL")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("storage.s3.token", "WATCHTOWER__STORAGE__S3__TOKEN")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}

	// Cache redis config
	bindErr = viperInstance.BindEnv("task.storage.redis.address", "WATCHTOWER__TASK__STORAGE__REDIS__ADDRESS")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("task.storage.redis.username", "WATCHTOWER__TASK__STORAGE__REDIS__USERNAME")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("task.storage.redis.password", "WATCHTOWER__TASK__STORAGE__REDIS__PASSWORD")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("task.storage.redis.expired", "WATCHTOWER__TASK__STORAGE__REDIS__EXPIRED")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}

	// Queue emq config
	bindErr = viperInstance.BindEnv("task.queue.rmq.address", "WATCHTOWER__TASK__QUEUE__RMQ__ADDRESS")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("task.queue.rmq.exchange", "WATCHTOWER__TASK__QUEUE__RMQ__EXCHANGE")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("task.queue.rmq.routing_key", "WATCHTOWER__TASK__QUEUE__RMQ__ROUTING_KEY")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("task.queue.rmq.queue", "WATCHTOWER__TASK__QUEUE__RMQ__QUEUE")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}

	// Docstorage config
	bindErr = viperInstance.BindEnv(
		"task.processor.docstorage.address",
		"WATCHTOWER__TASK__PROCESSOR__DOCSTORAGE__ADDRESS",
	)
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv(
		"task.processor.docstorage.timeout",
		"WATCHTOWER__TASK__PROCESSOR__DOCSTORAGE__TIMEOUT",
	)
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}

	// Docparser config
	bindErr = viperInstance.BindEnv(
		"task.processor.docparser.address",
		"WATCHTOWER__TASK__PROCESSOR__DOCPARSER__ADDRESS",
	)
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv(
		"task.processor.docparser.timeout",
		"WATCHTOWER__TASK__PROCESSOR__DOCPARSER__TIMEOUT",
	)
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}

	if err := viperInstance.ReadInConfig(); err != nil {
		confErr := fmt.Errorf("failed while reading config file %s: %w", filePath, err)
		return config, confErr
	}

	if err := viperInstance.Unmarshal(config); err != nil {
		confErr := fmt.Errorf("failed while unmarshaling config file %s: %w", filePath, err)
		return config, confErr
	}

	return config, nil
}
