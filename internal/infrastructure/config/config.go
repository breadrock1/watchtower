package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"watchtower/internal/infrastructure/dedoc"
	"watchtower/internal/infrastructure/doc-searcher"
	"watchtower/internal/infrastructure/httpserver"
	"watchtower/internal/infrastructure/pg"
	"watchtower/internal/infrastructure/redis"
	"watchtower/internal/infrastructure/rmq"
	"watchtower/internal/infrastructure/s3"
)

type Config struct {
	Ocr     OcrConfig     `mapstructure:"ocr"`
	Cacher  CacherConfig  `mapstructure:"cacher"`
	Storage StorageConfig `mapstructure:"storage"`
	Server  ServerConfig  `mapstructure:"server"`
	Cloud   CloudConfig   `mapstructure:"cloud"`
	Queue   QueueConfig   `mapstructure:"queue"`
	Watcher WatcherConfig `mapstructure:"watcher"`
}

type OcrConfig struct {
	Dedoc dedoc.Config `mapstructure:"dedoc"`
}

type CacherConfig struct {
	Redis redis.Config `mapstructure:"redis"`
}

type StorageConfig struct {
	Docsearcher doc_searcher.Config `mapstructure:"docsearcher"`
}

type ServerConfig struct {
	Http httpserver.Config `mapstructure:"http"`
}

type CloudConfig struct {
	Minio s3.Config `mapstructure:"s3"`
}

type QueueConfig struct {
	Rmq rmq.Config `mapstructure:"rmq"`
}

type WatcherConfig struct {
	Storage WatcherStorageConfig `mapstructure:"storage"`
}

type WatcherStorageConfig struct {
	Pg pg.Config `mapstructure:"pg"`
}

func FromFile(filePath string) (*Config, error) {
	_ = godotenv.Load()

	config := &Config{}

	viperInstance := viper.New()
	//viperInstance.SetConfigFile(filePath)
	//viperInstance.SetConfigType("toml")

	viperInstance.AutomaticEnv()
	viperInstance.SetEnvPrefix("watchtower")
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))

	// Http server config
	err := viperInstance.BindEnv("server.http.address", "WATCHTOWER__SERVER__HTTP__ADDRESS")
	err = viperInstance.BindEnv("server.http.logger.level", "WATCHTOWER__SERVER__HTTP__LOGGER__LEVEL")
	err = viperInstance.BindEnv("server.http.logger.address", "WATCHTOWER__SERVER__HTTP__LOGGER__ADDRESS")
	err = viperInstance.BindEnv("server.http.logger.enable_loki", "WATCHTOWER__SERVER__HTTP__LOGGER__ENABLE_LOKI")

	// OCR config
	err = viperInstance.BindEnv("ocr.dedoc.address", "WATCHTOWER__OCR__DEDOC__ADDRESS")
	err = viperInstance.BindEnv("ocr.dedoc.timeout", "WATCHTOWER__OCR__DEDOC__TIMEOUT")

	// Storage doc-searcher config
	err = viperInstance.BindEnv("storage.docsearcher.address", "WATCHTOWER__STORAGE__DOC_SEARCHER__ADDRESS")

	// Pg watched dirs config
	err = viperInstance.BindEnv("watcher.storage.pg.host", "WATCHTOWER__WATCHER__STORAGE__PG__HOST")
	err = viperInstance.BindEnv("watcher.storage.pg.port", "WATCHTOWER__WATCHER__STORAGE__PG__PORT")
	err = viperInstance.BindEnv("watcher.storage.pg.username", "WATCHTOWER__WATCHER__STORAGE__PG__USER")
	err = viperInstance.BindEnv("watcher.storage.pg.password", "WATCHTOWER__WATCHER__STORAGE__PG__PASSWORD")
	err = viperInstance.BindEnv("watcher.storage.pg.dbname", "WATCHTOWER__WATCHER__STORAGE__PG__DBNAME")
	err = viperInstance.BindEnv("watcher.storage.pg.ssl_mode", "WATCHTOWER__WATCHER__STORAGE__PG__SSL_MODE")

	// Cache redis config
	err = viperInstance.BindEnv("cacher.redis.address", "WATCHTOWER__CACHER__REDIS__ADDRESS")
	err = viperInstance.BindEnv("cacher.redis.username", "WATCHTOWER__CACHER__REDIS__USERNAME")
	err = viperInstance.BindEnv("cacher.redis.password", "WATCHTOWER__CACHER__REDIS__PASSWORD")
	err = viperInstance.BindEnv("cacher.redis.expired", "WATCHTOWER__CACHER__REDIS__EXPIRED")

	// Queue emq config
	err = viperInstance.BindEnv("queue.rmq.address", "WATCHTOWER__QUEUE__RMQ__ADDRESS")
	err = viperInstance.BindEnv("queue.rmq.exchange", "WATCHTOWER__QUEUE__RMQ__EXCHANGE")
	err = viperInstance.BindEnv("queue.rmq.routing_key", "WATCHTOWER__QUEUE__RMQ__ROUTING_KEY")
	err = viperInstance.BindEnv("queue.rmq.queue", "WATCHTOWER__QUEUE__RMQ__QUEUE")

	// MinIO config
	err = viperInstance.BindEnv("cloud.s3.address", "WATCHTOWER__CLOUD__S3__ADDRESS")
	err = viperInstance.BindEnv("cloud.s3.access_id", "WATCHTOWER__CLOUD__S3__ACCESS_ID")
	err = viperInstance.BindEnv("cloud.s3.secret_key", "WATCHTOWER__CLOUD__S3__SECRET_KEY")
	err = viperInstance.BindEnv("cloud.s3.enable_ssl", "WATCHTOWER__CLOUD__S3__ENABLE_SSL")
	err = viperInstance.BindEnv("cloud.s3.token", "WATCHTOWER__CLOUD__S3__TOKEN")
	err = viperInstance.BindEnv("cloud.s3.watched_dirs", "WATCHTOWER__CLOUD__S3__WATCHED_DIRS")

	if err != nil {
		confErr := fmt.Errorf("failed while binding env vars: %w", err)
		return config, confErr
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
