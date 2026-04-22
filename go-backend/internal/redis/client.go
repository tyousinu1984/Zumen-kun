package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/zumen-kun/go-backend/internal/config"
)

type Client struct {
	rdb *redis.Client
}

func NewClient(cfg *config.Config) (*Client, error) {
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	rdb := redis.NewClient(opt)

	// Test connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

func (c *Client) GetClient() *redis.Client {
	return c.rdb
}
