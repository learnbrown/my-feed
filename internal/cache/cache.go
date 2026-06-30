package cache

import (
	"context"
	"fmt"
	"time"
)

func (c *Client) GetBytes(ctx context.Context, key string) ([]byte, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("redis client is unavaliable")
	}
	return c.rdb.Get(ctx, key).Bytes()
}

func (c *Client) SetBytes(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if !c.Enabled() {
		return fmt.Errorf("redis client is unavaliable")
	}
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *Client) Del(ctx context.Context, key string) error {
	if !c.Enabled() {
		return fmt.Errorf("redis client is unavaliable")
	}
	return c.rdb.Del(ctx, key).Err()
}
