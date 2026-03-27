package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

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

const (
	launchModeEnvKey  = "WATCHTOWER__RUN_MODE"
	defaultLaunchMode = "development"
	serviceEnvPrefix  = "WATCHTOWER"
)

func InitConfig() (*Config, error) {
	launchMode := os.Getenv(launchModeEnvKey)
	if launchMode == "" {
		launchMode = defaultLaunchMode
	}

	viperInst := viper.New()

	viperInst.SetConfigName(launchMode)
	viperInst.SetConfigType("toml")

	viperInst.AddConfigPath(".")
	viperInst.AddConfigPath("./configs")

	if err := viperInst.ReadInConfig(); err != nil {
		//nolint
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Info("config file not found, using env vars")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	setupEnv(viperInst)

	config := &Config{}
	if err := viperInst.Unmarshal(config); err != nil {
		confErr := fmt.Errorf("failed while unmarshaling config: %w", err)
		return config, confErr
	}

	return config, nil
}

func setupEnv(viperInst *viper.Viper) {
	viperInst.AutomaticEnv()
	viperInst.SetEnvPrefix(serviceEnvPrefix)
	viperInst.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))

	//nolint
	envMappings := map[string]string{
		"orchestrator.semaphore_size":       "ORCHESTRATOR__SEMAPHORE_SIZE",
		"otlp.logger.level":                 "OTLP__LOGGER__LEVEL",
		"otlp.logger.address":               "OTLP__LOGGER__ADDRESS",
		"otlp.logger.enable_loki":           "OTLP__LOGGER__ENABLE_LOKI",
		"otlp.tracer.address":               "OTLP__TRACER__ADDRESS",
		"otlp.tracer.enable_jaeger":         "OTLP__TRACER__ENABLE_JAEGER",
		"server.http.address":               "SERVER__HTTP__ADDRESS",
		"storage.s3.address":                "STORAGE__S3__ADDRESS",
		"storage.s3.access_id":              "STORAGE__S3__ACCESS_ID",
		"storage.s3.secret_key":             "STORAGE__S3__SECRET_KEY",
		"storage.s3.enable_ssl":             "STORAGE__S3__ENABLE_SSL",
		"storage.s3.token":                  "STORAGE__S3__TOKEN",
		"task.storage.redis.address":        "TASK__STORAGE__REDIS__ADDRESS",
		"task.storage.redis.username":       "TASK__STORAGE__REDIS__USERNAME",
		"task.storage.redis.password":       "TASK__STORAGE__REDIS__PASSWORD",
		"task.storage.redis.expired":        "TASK__STORAGE__REDIS__EXPIRED",
		"task.queue.rmq.address":            "TASK__QUEUE__RMQ__ADDRESS",
		"task.queue.rmq.exchange":           "TASK__QUEUE__RMQ__EXCHANGE",
		"task.queue.rmq.routing_key":        "TASK__QUEUE__RMQ__ROUTING_KEY",
		"task.queue.rmq.queue":              "TASK__QUEUE__RMQ__QUEUE",
		"task.processor.docstorage.address": "TASK__PROCESSOR__DOCSTORAGE__ADDRESS",
		"task.processor.docstorage.timeout": "TASK__PROCESSOR__DOCSTORAGE__TIMEOUT",
		"task.processor.docparser.address":  "TASK__PROCESSOR__DOCPARSER__ADDRESS",
		"task.processor.docparser.timeout":  "TASK__PROCESSOR__DOCPARSER__TIMEOUT",
	}

	var bindErr error
	for key, value := range envMappings {
		bindErr = viperInst.BindEnv(key, fmt.Sprintf("%s__%s", serviceEnvPrefix, value))
		if bindErr != nil {
			slog.Warn("failed to bind env var", slog.String("err", bindErr.Error()))
		}
	}
}
