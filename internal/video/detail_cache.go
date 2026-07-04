package video

import (
	"context"
	"encoding/json"
	"errors"
	"my_feed/internal/cache"
	"time"

	"github.com/redis/go-redis/v9"
)

type DetailCache interface {
	GetDetail(ctx context.Context, videoID uint) (VideoDTO, bool, error)
	SetDetail(ctx context.Context, videoID uint, video VideoDTO, ttl time.Duration) error
	DelDetail(ctx context.Context, videoID uint) error
}

type RedisDetailCache struct {
	cache *cache.Client
}

func NewRedisDetailCache(cache *cache.Client) *RedisDetailCache {
	return &RedisDetailCache{cache: cache}
}

func (dc *RedisDetailCache) GetDetail(ctx context.Context, videoID uint) (VideoDTO, bool, error) {
	if dc == nil || dc.cache == nil || !dc.cache.Enabled() {
		return VideoDTO{}, false, cache.ErrDisabled
	}
	b, err := dc.cache.GetBytes(ctx, dc.cache.VideoDetailKey(videoID))
	if errors.Is(err, redis.Nil) {
		return VideoDTO{}, false, nil
	}
	if err != nil {
		return VideoDTO{}, false, err
	}
	detail := &VideoDTO{}
	err = json.Unmarshal(b, detail)
	if err != nil {
		// 反序列化失败则删除错误缓存
		dc.cache.Del(ctx, dc.cache.VideoDetailKey(videoID))
		return VideoDTO{}, false, nil
	}
	return *detail, true, nil
}

func (dc *RedisDetailCache) SetDetail(ctx context.Context, videoID uint, video VideoDTO, ttl time.Duration) error {
	if dc == nil || dc.cache == nil || !dc.cache.Enabled() {
		return cache.ErrDisabled
	}

	b, err := json.Marshal(&video)
	if err != nil {
		return err
	}
	return dc.cache.SetBytes(ctx, dc.cache.VideoDetailKey(videoID), b, ttl)
}

func (dc *RedisDetailCache) DelDetail(ctx context.Context, videoID uint) error {
	if dc == nil || dc.cache == nil || !dc.cache.Enabled() {
		return cache.ErrDisabled
	}
	return dc.cache.Del(ctx, dc.cache.VideoDetailKey(videoID))
}
