package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"watchtower/internal/shared/kernel"
	"watchtower/internal/shared/telemetry"
	"watchtower/internal/support/task/domain"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	config Config
	rsConn *redis.Client
}

func New(config Config) domain.ITaskStorage {
	redisOpts := &redis.Options{Addr: config.Address}
	conn := redis.NewClient(redisOpts)
	return &RedisClient{
		config: config,
		rsConn: conn,
	}
}

func (rs *RedisClient) GetAllBucketTasks(ctx context.Context, bucketID kernel.BucketID) ([]*domain.Task, error) {
	key := rs.generateUniqID(bucketID, "*")
	status := rs.rsConn.Scan(ctx, 0, key, -1)
	if status.Err() != nil {
		return nil, fmt.Errorf("redis error: %w", status.Err())
	}

	rKeys, _ := status.Val()
	tasks := make([]*domain.Task, len(rKeys))
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

		task, err := value.ConvertToTask()
		if err != nil {
			slog.Warn("failed to unmarshal task", slog.String("err", err.Error()))
			continue
		}

		tasks[index] = task
	}

	return tasks, nil
}

func (rs *RedisClient) GetTask(
	ctx context.Context,
	bucketID kernel.BucketID,
	taskID domain.TaskID,
) (*domain.Task, error) {
	key := rs.generateUniqID(bucketID, taskID.String())
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

	taskEvent, err := value.ConvertToTask()
	if err != nil {
		return nil, fmt.Errorf("task validation error: %w", err)
	}

	return taskEvent, nil
}

func (rs *RedisClient) UpdateTask(ctx context.Context, task *domain.Task) error {
	key := rs.generateUniqID(task.BucketID, task.ID.String())

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

func (rs *RedisClient) generateUniqID(bucketID kernel.BucketID, taskID string) string {
	return fmt.Sprintf("%s:%s:%s", telemetry.AppName, bucketID, taskID)
}
