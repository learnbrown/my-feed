package cache

import (
	"context"
	"fmt"
	"my_feed/internal/config"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb       *redis.Client
	keyPrefix string
	enabled   bool
}

const defaultPrefix = "myfeed"

func NewRedis(cfg *config.RedisConfig) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("redis config required")
	}
	if !cfg.Enabled {
		return nil, fmt.Errorf("config disabled redis")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	client := &Client{rdb: rdb, keyPrefix: cfg.KeyPrefix, enabled: cfg.Enabled}
	if client.keyPrefix == "" {
		client.keyPrefix = defaultPrefix
	}

	return client, nil
}

func (c *Client) Close() error {
	if !c.Enabled() {
		return nil
	}
	return c.rdb.Close()
}

func (c *Client) Ping(ctx context.Context) error {
	if !c.Enabled() {
		return ErrDisabled
	}
	return c.rdb.Ping(ctx).Err()
}

func (c *Client) Enabled() bool {
	if c == nil || c.rdb == nil {
		return false
	}
	return c.enabled
}
