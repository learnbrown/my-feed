package account

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

func TestRedisTokenCacheDisabled(t *testing.T) {
	tokenCache := NewRedisTokenCache(nil)

	if _, _, err := tokenCache.GetToken(t.Context(), 1); !errors.Is(err, cache.ErrDisabled) {
		t.Fatalf("GetToken() error = %v, want ErrDisabled", err)
	}
	if err := tokenCache.SetToken(t.Context(), 1, "token", time.Minute); !errors.Is(err, cache.ErrDisabled) {
		t.Fatalf("SetToken() error = %v, want ErrDisabled", err)
	}
	if err := tokenCache.DelToken(t.Context(), 1); !errors.Is(err, cache.ErrDisabled) {
		t.Fatalf("DelToken() error = %v, want ErrDisabled", err)
	}
}

func TestCalculateTokenCacheTTL(t *testing.T) {
	now := time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		expiresAt time.Time
		want      time.Duration
	}{
		{name: "caps long JWT lifetime", expiresAt: now.Add(time.Hour), want: MaxTokenCacheTTL},
		{name: "uses short remaining lifetime", expiresAt: now.Add(90 * time.Second), want: 90 * time.Second},
		{name: "expired token", expiresAt: now.Add(-time.Second), want: 0},
		{name: "expires now", expiresAt: now, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateTokenCacheTTL(tt.expiresAt, now); got != tt.want {
				t.Fatalf("CalculateTokenCacheTTL() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRedisTokenCacheLifecycle(t *testing.T) {
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
	tokenCache := NewRedisTokenCache(client)

	if token, hit, err := tokenCache.GetToken(t.Context(), 7); err != nil || hit || token != "" {
		t.Fatalf("initial GetToken() = (%q, %t, %v), want miss", token, hit, err)
	}
	if err := tokenCache.SetToken(t.Context(), 7, "token-value", time.Minute); err != nil {
		t.Fatalf("SetToken() error = %v", err)
	}
	if token, hit, err := tokenCache.GetToken(t.Context(), 7); err != nil || !hit || token != "token-value" {
		t.Fatalf("GetToken() = (%q, %t, %v)", token, hit, err)
	}
	if ttl := mini.TTL("test:account:token:7"); ttl != time.Minute {
		t.Fatalf("token TTL = %s, want 1m", ttl)
	}

	mini.FastForward(20 * time.Second)
	if token, hit, err := tokenCache.GetToken(t.Context(), 7); err != nil || !hit || token != "token-value" {
		t.Fatalf("GetToken() after time advance = (%q, %t, %v)", token, hit, err)
	}
	if ttl := mini.TTL("test:account:token:7"); ttl != 40*time.Second {
		t.Fatalf("token TTL after cache hit = %s, want 40s", ttl)
	}

	mini.FastForward(41 * time.Second)
	if token, hit, err := tokenCache.GetToken(t.Context(), 7); err != nil || hit || token != "" {
		t.Fatalf("expired GetToken() = (%q, %t, %v), want miss", token, hit, err)
	}

	if err := tokenCache.SetToken(t.Context(), 7, "expired-token", 0); err != nil {
		t.Fatalf("SetToken() with zero TTL error = %v", err)
	}
	if mini.Exists("test:account:token:7") {
		t.Fatal("zero TTL token should not be cached")
	}

	if err := tokenCache.SetToken(t.Context(), 7, "new-token", time.Minute); err != nil {
		t.Fatalf("second SetToken() error = %v", err)
	}
	if err := tokenCache.DelToken(t.Context(), 7); err != nil {
		t.Fatalf("DelToken() error = %v", err)
	}
	if mini.Exists("test:account:token:7") {
		t.Fatal("token key still exists after deletion")
	}
}
