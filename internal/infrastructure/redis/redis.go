package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"watchtower/internal/application/models"
)

type RedisClient struct {
	config *Config
	rsConn *redis.Client
}

func New(config *Config) *RedisClient {
	redisOpts := &redis.Options{Addr: config.Address}
	conn := redis.NewClient(redisOpts)
	return &RedisClient{
		config: config,
		rsConn: conn,
	}
}

func (rs *RedisClient) GetAll(ctx context.Context, bucket string) ([]*models.TaskEvent, error) {
	key := rs.generateUniqID(bucket, "*")
	status := rs.rsConn.Scan(ctx, 0, key, -1)
	if status.Err() != nil {
		return nil, fmt.Errorf("redis error: %w", status.Err())
	}

	rKeys, _ := status.Val()
	tasks := make([]*models.TaskEvent, len(rKeys))
	for index, rKey := range rKeys {
		cmd := rs.rsConn.Get(ctx, rKey)
		data, err := cmd.Bytes()
		if err != nil {
			slog.Warn("failed to get task", slog.String("err", err.Error()))
			continue
		}

		value := &RedisValue{}
		if err = json.Unmarshal(data, &value); err != nil {
			slog.Warn("failed to unmarshal task", slog.String("err", err.Error()))
			continue
		}

		taskEvent, err := value.ConvertToTaskEvent()
		if err != nil {
			slog.Warn("failed to unmarshal task", slog.String("err", err.Error()))
			continue
		}

		tasks[index] = taskEvent
	}

	return tasks, nil
}

func (rs *RedisClient) Get(ctx context.Context, bucket, taskID string) (*models.TaskEvent, error) {
	key := rs.generateUniqID(bucket, taskID)
	cmd := rs.rsConn.Get(ctx, key)
	if cmd.Err() != nil {
		return nil, fmt.Errorf("redis error: %w", cmd.Err())
	}

	data, err := cmd.Bytes()
	if err != nil {
		return nil, fmt.Errorf("read bytes data error: %w", err)
	}

	value := &RedisValue{}
	if err = json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("deserialize error: %w", err)
	}

	taskEvent, err := value.ConvertToTaskEvent()
	if err != nil {
		return nil, fmt.Errorf("task validation error: %w", err)
	}

	return taskEvent, nil
}

func (rs *RedisClient) Push(ctx context.Context, task *models.TaskEvent) error {
	key := rs.generateUniqID(task.Bucket, task.ID.String())

	value := ConvertFromTaskEvent(task)
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("serialize error: %w", err)
	}

	status := rs.rsConn.Set(ctx, key, jsonData, rs.config.Expired*time.Second)
	if status.Err() != nil {
		return fmt.Errorf("redis error: %w", status.Err())
	}

	return nil
}

func (rs *RedisClient) generateUniqID(bucket, taskID string) string {
	return fmt.Sprintf("watchtower:%s:%s", bucket, taskID)
}
