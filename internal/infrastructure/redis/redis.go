package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"watchtower/internal/application/dto"
	"watchtower/internal/application/utils"
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

func (rs *RedisClient) GetAll(ctx context.Context, bucket string) ([]*dto.TaskEvent, error) {
	key := rs.buildKey(bucket, "*")
	status := rs.rsConn.Scan(ctx, 0, key, -1)
	if status.Err() != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", status.Err())
	}

	rKeys, _ := status.Val()
	tasks := make([]*dto.TaskEvent, len(rKeys))
	for index, rKey := range rKeys {
		cmd := rs.rsConn.Get(ctx, rKey)
		data, err := cmd.Bytes()
		if err != nil {
			log.Printf("failed to get task: %w", err)
			continue
		}

		value := &RedisValue{}
		if err = json.Unmarshal(data, &value); err != nil {
			log.Printf("failed to unmarshal task: %w", err)
			continue
		}

		taskEvent, err := value.ConvertToTaskEvent()
		if err != nil {
			log.Printf("failed to unmarshal task: %w", err)
			continue
		}

		tasks[index] = taskEvent
	}

	return tasks, nil
}

func (rs *RedisClient) Get(ctx context.Context, bucket, file string) (*dto.TaskEvent, error) {
	key := rs.buildKey(bucket, file)
	cmd := rs.rsConn.Get(ctx, key)
	if cmd.Err() != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", cmd.Err())
	}

	data, err := cmd.Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	value := &RedisValue{}
	if err = json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	taskEvent, err := value.ConvertToTaskEvent()
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return taskEvent, nil
}

func (rs *RedisClient) Push(ctx context.Context, task *dto.TaskEvent) error {
	key := rs.buildKey(task.Bucket, task.FilePath)

	value := ConvertFromTaskEvent(task)
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	status := rs.rsConn.Set(ctx, key, jsonData, rs.config.Expired*time.Second)
	if status.Err() != nil {
		return fmt.Errorf("redis set task status failed, %w", status.Err())
	}

	return nil
}

func (rs *RedisClient) buildKey(bucket, suffix string) string {
	if suffix != "*" {
		suffix = utils.ComputeMd5([]byte(suffix))
	}

	return fmt.Sprintf("%s:%s", bucket, suffix)
}
