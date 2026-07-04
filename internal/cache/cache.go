package cache

import (
	"context"
	"errors"
	"time"
)

var ErrDisabled = errors.New("redis disabled")

func (c *Client) GetBytes(ctx context.Context, key string) ([]byte, error) {
	if !c.Enabled() {
		return nil, ErrDisabled
	}
	return c.rdb.Get(ctx, key).Bytes()
}

func (c *Client) SetBytes(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if !c.Enabled() {
		return ErrDisabled
	}
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *Client) Del(ctx context.Context, key string) error {
	if !c.Enabled() {
		return ErrDisabled
	}
	return c.rdb.Del(ctx, key).Err()
}
