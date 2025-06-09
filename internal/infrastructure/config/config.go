package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"watchtower/internal/infrastructure/dedoc"
	"watchtower/internal/infrastructure/doc-searcher"
	"watchtower/internal/infrastructure/httpserver"
	"watchtower/internal/infrastructure/redis"
	"watchtower/internal/infrastructure/rmq"
	"watchtower/internal/infrastructure/s3"
	"watchtower/internal/infrastructure/vectorizer"
)

type Config struct {
	Ocr       OcrConfig       `mapstructure:"ocr"`
	Cacher    CacherConfig    `mapstructure:"cacher"`
	Storage   StorageConfig   `mapstructure:"storage"`
	Server    ServerConfig    `mapstructure:"server"`
	Tokenizer TokenizerConfig `mapstructure:"tokenizer"`
	Cloud     CloudConfig     `mapstructure:"cloud"`
	Queue     QueueConfig     `mapstructure:"queue"`
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

type TokenizerConfig struct {
	Vectorizer vectorizer.Config `mapstructure:"vectorizer"`
}

type CloudConfig struct {
	Minio s3.Config `mapstructure:"minio"`
}

type QueueConfig struct {
	Rmq rmq.Config `mapstructure:"rmq"`
}

func FromFile(filePath string) (*Config, error) {
	_ = godotenv.Load()

	config := &Config{}

	viperInstance := viper.New()
	viperInstance.SetConfigFile(filePath)
	viperInstance.SetConfigType("toml")

	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	viperInstance.AutomaticEnv()

	err := viperInstance.BindEnv(
		"server.http.address",
		"server.http.logger.level",
		"server.http.logger.enableloki",
		"server.http.logger.address",

		"ocr.dedoc.address",
		"ocr.dedoc.timeout",
		"ocr.dedoc.enable_ssl",

		"storage.doc_searcher.address",
		"storage.doc_searcher.enable_ssl",

		"cacher.redis.address",
		"cacher.redis.username",
		"cacher.redis.password",
		"cacher.redis.expired",

		"queue.rmq.address",
		"queue.rmq.exchange",
		"queue.rmq.routing_key",
		"queue.rmq.queue_name",

		"cloud.minio.address",
		"cloud.minio.access_id",
		"cloud.minio.secret_key",
		"cloud.minio.enable_ssl",
		"cloud.minio.token",

		"tokenizer.vectorizer.address",
		"tokenizer.vectorizer.enable_ssl",
		"tokenizer.vectorizer.chunk_size",
		"tokenizer.vectorizer.chunk_overlap",
		"tokenizer.vectorizer.return_chunks",
		"tokenizer.vectorizer.chunks_by_self",
	)

	if err != nil {
		confErr := fmt.Errorf("failed while binding env vars: %v", err)
		return config, confErr
	}

	if err := viperInstance.ReadInConfig(); err != nil {
		confErr := fmt.Errorf("failed while reading config file %s: %v", filePath, err)
		return config, confErr
	}

	if err := viperInstance.Unmarshal(config); err != nil {
		confErr := fmt.Errorf("failed while unmarshaling config file %s: %v", filePath, err)
		return config, confErr
	}

	return config, nil
}
