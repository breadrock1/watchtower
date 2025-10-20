package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"watchtower/internal/application/utils/telemetry"
	"watchtower/internal/infrastructure/doc-parser"
	"watchtower/internal/infrastructure/doc-storage"
	"watchtower/internal/infrastructure/redis"
	"watchtower/internal/infrastructure/rmq"
	"watchtower/internal/infrastructure/s3"
)

type Config struct {
	Settings   SettingsConfig   `mapstructure:"settings"`
	Server     ServerConfig     `mapstructure:"server"`
	Recognizer RecognizerConfig `mapstructure:"recognizer"`
	Task       TaskConfig       `mapstructure:"task"`
	Storage    StorageConfig    `mapstructure:"storage"`
}

type SettingsConfig struct {
	PipelineSemaphoreSize int `mapstructure:"pipeline_semaphore_size"`
}

type ServerConfig struct {
	Http   HttpServerConfig       `mapstructure:"http"`
	Logger telemetry.LoggerConfig `mapstructure:"logger"`
	Tracer telemetry.TracerConfig `mapstructure:"tracer"`
}

type HttpServerConfig struct {
	Address string `mapstructure:"address"`
}

type RecognizerConfig struct {
	DocParser doc_parser.Config `mapstructure:"docparser"`
}

type TaskConfig struct {
	TaskStorage TaskStorageConfig `mapstructure:"storage"`
	TaskQueue   TaskQueueConfig   `mapstructure:"queue"`
}

type StorageConfig struct {
	ObjectStorage   ObjectStorageConfig   `mapstructure:"object"`
	DocumentStorage DocumentStorageConfig `mapstructure:"document"`
}

type TaskStorageConfig struct {
	Redis redis.Config `mapstructure:"redis"`
}

type TaskQueueConfig struct {
	Rmq rmq.Config `mapstructure:"rmq"`
}

type ObjectStorageConfig struct {
	S3 s3.Config `mapstructure:"s3"`
}

type DocumentStorageConfig struct {
	DocSearcher doc_storage.Config `mapstructure:"docsearcher"`
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

	// Settings config
	bindErr := viperInstance.BindEnv(
		"settings.pipeline_semaphore_size",
		"WATCHTOWER__SETTINGS__PIPELINE_SEMAPHORE_SIZE",
	)
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}

	// Server config
	bindErr = viperInstance.BindEnv("server.http.address", "WATCHTOWER__SERVER__HTTP__ADDRESS")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}

	bindErr = viperInstance.BindEnv("server.logger.level", "WATCHTOWER__SERVER__LOGGER__LEVEL")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("server.logger.address", "WATCHTOWER__SERVER__LOGGER__ADDRESS")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("server.logger.enable_loki", "WATCHTOWER__SERVER__LOGGER__ENABLE_LOKI")

	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("server.tracer.address", "WATCHTOWER__SERVER__TRACER__ADDRESS")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("server.tracer.enable_jaeger", "WATCHTOWER__SERVER__TRACER__ENABLE_JAEGER")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}

	// Recognizer config
	bindErr = viperInstance.BindEnv("recognizer.docparser.address", "WATCHTOWER__RECOGNIZER__DOCPARSER__ADDRESS")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("recognizer.docparser.timeout", "WATCHTOWER__RECOGNIZER__DOCPARSER__TIMEOUT")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}

	// Storage config
	bindErr = viperInstance.BindEnv(
		"storage.document.docsearcher.address",
		"WATCHTOWER__STORAGE__DOCUMENT__DOCSEARCHER__ADDRESS",
	)
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv(
		"storage.document.docsearcher.timeout",
		"WATCHTOWER__STORAGE__DOCUMENT__DOCSEARCHER__TIMEOUT",
	)
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

	// Storage s3 config
	bindErr = viperInstance.BindEnv(
		"storage.object.s3.address",
		"WATCHTOWER__STORAGE__OBJECT__CLOUD__S3__ADDRESS",
	)
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv(
		"storage.object.s3.access_id",
		"WATCHTOWER__STORAGE__OBJECT__CLOUD__S3__ACCESS_ID",
	)
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv(
		"storage.object.s3.secret_key",
		"WATCHTOWER__STORAGE__OBJECT__CLOUD__S3__SECRET_KEY",
	)
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv(
		"storage.object.s3.enable_ssl",
		"WATCHTOWER__STORAGE__OBJECT__CLOUD__S3__ENABLE_SSL",
	)
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv(
		"storage.object.s3.token",
		"WATCHTOWER__STORAGE__OBJECT__CLOUD__S3__TOKEN",
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
