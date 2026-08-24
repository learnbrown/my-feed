package video

import (
	"errors"
	"net"
	"strconv"
	"testing"
	"time"

	"my_feed/internal/cache"
	"my_feed/internal/config"

	"github.com/alicebob/miniredis/v2"
)

func TestRedisDetailCacheDisabled(t *testing.T) {
	detailCache := NewRedisDetailCache(nil)

	if _, _, err := detailCache.GetDetail(t.Context(), 1); !errors.Is(err, cache.ErrDisabled) {
		t.Fatalf("GetDetail() error = %v, want ErrDisabled", err)
	}
	if err := detailCache.SetDetail(t.Context(), 1, VideoDTO{}, time.Minute); !errors.Is(err, cache.ErrDisabled) {
		t.Fatalf("SetDetail() error = %v, want ErrDisabled", err)
	}
	if err := detailCache.DelDetail(t.Context(), 1); !errors.Is(err, cache.ErrDisabled) {
		t.Fatalf("DelDetail() error = %v, want ErrDisabled", err)
	}
}

func TestRedisDetailCacheLifecycleAndCorruptValueCleanup(t *testing.T) {
	mini := miniredis.RunT(t)
	host, portString, err := net.SplitHostPort(mini.Addr())
	if err != nil {
		t.Fatalf("split miniredis address: %v", err)
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		t.Fatalf("parse miniredis port: %v", err)
	}
	client, err := cache.NewRedis(&config.RedisConfig{Host: host, Port: port, KeyPrefix: "test", Enabled: true})
	if err != nil {
		t.Fatalf("NewRedis() error = %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	detailCache := NewRedisDetailCache(client)
	want := VideoDTO{ID: 9, AuthorID: 2, Title: "cached", CreatedAt: 1_700_000_000_123}

	if got, hit, err := detailCache.GetDetail(t.Context(), want.ID); err != nil || hit || got != (VideoDTO{}) {
		t.Fatalf("initial GetDetail() = (%#v, %t, %v), want miss", got, hit, err)
	}
	if err := detailCache.SetDetail(t.Context(), want.ID, want, 5*time.Minute); err != nil {
		t.Fatalf("SetDetail() error = %v", err)
	}
	if got, hit, err := detailCache.GetDetail(t.Context(), want.ID); err != nil || !hit || got != want {
		t.Fatalf("GetDetail() = (%#v, %t, %v), want %#v", got, hit, err, want)
	}
	key := client.VideoDetailKey(want.ID)
	if ttl := mini.TTL(key); ttl != 5*time.Minute {
		t.Fatalf("detail TTL = %s, want 5m", ttl)
	}

	if err := client.SetBytes(t.Context(), key, []byte("not-json"), time.Minute); err != nil {
		t.Fatalf("write corrupt cache value: %v", err)
	}
	if got, hit, err := detailCache.GetDetail(t.Context(), want.ID); err != nil || hit || got != (VideoDTO{}) {
		t.Fatalf("corrupt GetDetail() = (%#v, %t, %v), want miss", got, hit, err)
	}
	if mini.Exists(key) {
		t.Fatal("corrupt detail cache was not deleted")
	}
}
