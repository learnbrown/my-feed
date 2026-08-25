package profile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"my_feed/internal/cache"
	"time"

	"github.com/redis/go-redis/v9"
)

type ProfileCache interface {
	GetProfile(ctx context.Context, accountID uint) (Profile, bool, error)
	SetProfile(ctx context.Context, accountID uint, profile Profile, ttl time.Duration) error
	DelProfile(ctx context.Context, accountID uint) error
}

type RedisProfileCache struct {
	cache *cache.Client
}

func NewRedisProfileCache(cache *cache.Client) *RedisProfileCache {
	return &RedisProfileCache{cache: cache}
}

func (pc *RedisProfileCache) GetProfile(ctx context.Context, accountID uint) (Profile, bool, error) {
	if pc == nil || pc.cache == nil || !pc.cache.Enabled() {
		return Profile{}, false, cache.ErrDisabled
	}

	key := pc.cache.ProfileKey(accountID)
	b, err := pc.cache.GetBytes(ctx, key)
	if errors.Is(err, redis.Nil) {
		return Profile{}, false, nil
	}
	if err != nil {
		return Profile{}, false, err
	}

	profile := Profile{}
	if err := json.Unmarshal(b, &profile); err != nil {
		_ = pc.cache.Del(ctx, key)
		return Profile{}, false, fmt.Errorf("decode profile cache: %w", err)
	}

	return profile, true, nil
}

func (pc *RedisProfileCache) SetProfile(ctx context.Context, accountID uint, profile Profile, ttl time.Duration) error {
	if pc == nil || pc.cache == nil || !pc.cache.Enabled() {
		return cache.ErrDisabled
	}

	b, err := json.Marshal(&profile)
	if err != nil {
		return err
	}

	return pc.cache.SetBytes(ctx, pc.cache.ProfileKey(accountID), b, ttl)
}

func (pc *RedisProfileCache) DelProfile(ctx context.Context, accountID uint) error {
	if pc == nil || pc.cache == nil || !pc.cache.Enabled() {
		return cache.ErrDisabled
	}

	return pc.cache.Del(ctx, pc.cache.ProfileKey(accountID))
}
