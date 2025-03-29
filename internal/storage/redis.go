package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Rammurthy5/ai_agents_wrapper/internal/facade"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisClient(url string) (*RedisClient, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %v", err)
	}

	client := redis.NewClient(opt)
	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisClient{client: client, ctx: ctx}, nil
}

func (r *RedisClient) StoreResult(taskID string, result facade.MergedApiResponse) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal error: %v", err)
	}

	// Store with a TTL of 1 hour (adjust as needed)
	err = r.client.Set(r.ctx, taskID, data, 1*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to store result in Redis: %v", err)
	}
	return nil
}

func (r *RedisClient) GetResult(taskID string) (*facade.MergedApiResponse, error) {
	data, err := r.client.Get(r.ctx, taskID).Bytes()
	if err == redis.Nil {
		return nil, nil // Not found
	} else if err != nil {
		return nil, fmt.Errorf("failed to get result from Redis: %v", err)
	}

	var result facade.MergedApiResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal error: %v", err)
	}
	return &result, nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
