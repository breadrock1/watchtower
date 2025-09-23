package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"watchtower/internal/application/utils/telemetry"
	"watchtower/internal/infrastructure/dedoc"
	"watchtower/internal/infrastructure/doc-storage"
	"watchtower/internal/infrastructure/redis"
	"watchtower/internal/infrastructure/rmq"
	"watchtower/internal/infrastructure/s3"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Ocr        OcrConfig        `mapstructure:"ocr"`
	DocStorage DocStorageConfig `mapstructure:"storage"`
	Cacher     CacherConfig     `mapstructure:"cacher"`
	Queue      QueueConfig      `mapstructure:"queue"`
	Cloud      CloudConfig      `mapstructure:"cloud"`
	Settings   SettingsConfig   `mapstructure:"settings"`
}

type SettingsConfig struct {
	ChunkSize    int `mapstructure:"chunk_size"`
	ChunkOverlap int `mapstructure:"chunk_overlap"`
}

type ServerConfig struct {
	Http   HttpServerConfig       `mapstructure:"http"`
	Logger telemetry.LoggerConfig `mapstructure:"logger"`
	Tracer telemetry.TracerConfig `mapstructure:"tracer"`
}

type HttpServerConfig struct {
	Address string `mapstructure:"address"`
}

type OcrConfig struct {
	Dedoc dedoc.Config `mapstructure:"dedoc"`
}

type DocStorageConfig struct {
	DocSearcher doc_storage.Config `mapstructure:"docsearcher"`
}

type CacherConfig struct {
	Redis redis.Config `mapstructure:"redis"`
}

type QueueConfig struct {
	Rmq rmq.Config `mapstructure:"rmq"`
}

type CloudConfig struct {
	S3 s3.Config `mapstructure:"s3"`
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

	// Settings server config
	bindErr := viperInstance.BindEnv("settings.chunk_size", "WATCHTOWER__SETTINGS__CHUNK_SIZE")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("settings.chunk_overlap", "WATCHTOWER__SETTINGS__CHUNK_OVERLAP")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}

	// Http server config
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

	// OCR config
	bindErr = viperInstance.BindEnv("ocr.dedoc.address", "WATCHTOWER__OCR__DEDOC__ADDRESS")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("ocr.dedoc.timeout", "WATCHTOWER__OCR__DEDOC__TIMEOUT")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}

	// Storage doc-searcher config
	bindErr = viperInstance.BindEnv("storage.docsearcher.address", "WATCHTOWER__STORAGE__DOC_SEARCHER__ADDRESS")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}

	// Cache redis config
	bindErr = viperInstance.BindEnv("cacher.redis.address", "WATCHTOWER__CACHER__REDIS__ADDRESS")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("cacher.redis.username", "WATCHTOWER__CACHER__REDIS__USERNAME")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("cacher.redis.password", "WATCHTOWER__CACHER__REDIS__PASSWORD")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("cacher.redis.expired", "WATCHTOWER__CACHER__REDIS__EXPIRED")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}

	// Queue emq config
	bindErr = viperInstance.BindEnv("queue.rmq.address", "WATCHTOWER__QUEUE__RMQ__ADDRESS")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("queue.rmq.exchange", "WATCHTOWER__QUEUE__RMQ__EXCHANGE")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("queue.rmq.routing_key", "WATCHTOWER__QUEUE__RMQ__ROUTING_KEY")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("queue.rmq.queue", "WATCHTOWER__QUEUE__RMQ__QUEUE")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}

	// MinIO config
	bindErr = viperInstance.BindEnv("cloud.s3.address", "WATCHTOWER__CLOUD__S3__ADDRESS")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("cloud.s3.access_id", "WATCHTOWER__CLOUD__S3__ACCESS_ID")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("cloud.s3.secret_key", "WATCHTOWER__CLOUD__S3__SECRET_KEY")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("cloud.s3.enable_ssl", "WATCHTOWER__CLOUD__S3__ENABLE_SSL")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("cloud.s3.token", "WATCHTOWER__CLOUD__S3__TOKEN")
	if bindErr != nil {
		return nil, fmt.Errorf("failed to bine env varialbe: %w", bindErr)
	}
	bindErr = viperInstance.BindEnv("cloud.s3.watched_dirs", "WATCHTOWER__CLOUD__S3__WATCHED_DIRS")
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
