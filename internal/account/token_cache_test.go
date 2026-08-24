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

	mini.FastForward(time.Minute + time.Second)
	if token, hit, err := tokenCache.GetToken(t.Context(), 7); err != nil || hit || token != "" {
		t.Fatalf("expired GetToken() = (%q, %t, %v), want miss", token, hit, err)
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
