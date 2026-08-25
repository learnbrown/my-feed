package account

import (
	"context"
	"errors"
	"my_feed/internal/cache"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenCache interface {
	GetToken(ctx context.Context, accountID uint) (string, bool, error)
	SetToken(ctx context.Context, accountID uint, token string, ttl time.Duration) error
	DelToken(ctx context.Context, accountID uint) error
}

const MaxTokenCacheTTL = 5 * time.Minute

func CalculateTokenCacheTTL(expiresAt, now time.Time) time.Duration {
	remaining := expiresAt.Sub(now)
	if remaining <= 0 {
		return 0
	}
	if remaining > MaxTokenCacheTTL {
		return MaxTokenCacheTTL
	}
	return remaining
}

type RedisTokenCache struct {
	cache *cache.Client
}

func NewRedisTokenCache(cache *cache.Client) *RedisTokenCache {
	return &RedisTokenCache{cache: cache}
}

func (tc *RedisTokenCache) GetToken(ctx context.Context, accountID uint) (string, bool, error) {
	if tc == nil || tc.cache == nil || !tc.cache.Enabled() {
		return "", false, cache.ErrDisabled
	}
	key := tc.cache.AccountTokenKey(accountID)
	b, err := tc.cache.GetBytes(ctx, key)
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(b), true, nil
}

func (tc *RedisTokenCache) SetToken(ctx context.Context, accountID uint, token string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	if tc == nil || tc.cache == nil || !tc.cache.Enabled() {
		return cache.ErrDisabled
	}
	return tc.cache.SetBytes(ctx, tc.cache.AccountTokenKey(accountID), []byte(token), ttl)
}

func (tc *RedisTokenCache) DelToken(ctx context.Context, accountID uint) error {
	if tc == nil || tc.cache == nil || !tc.cache.Enabled() {
		return cache.ErrDisabled
	}
	return tc.cache.Del(ctx, tc.cache.AccountTokenKey(accountID))
}
