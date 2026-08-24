package cache

import (
	"testing"

	"my_feed/internal/config"
)

func TestNewRedisRejectsMissingOrDisabledConfig(t *testing.T) {
	if _, err := NewRedis(nil); err == nil {
		t.Fatal("expected nil config to be rejected")
	}
	if _, err := NewRedis(&config.RedisConfig{}); err == nil {
		t.Fatal("expected disabled config to be rejected")
	}
}

func TestNewRedisBuildsEnabledClientWithoutConnecting(t *testing.T) {
	client, err := NewRedis(&config.RedisConfig{
		Host:    "127.0.0.1",
		Port:    6379,
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("NewRedis() error = %v", err)
	}
	if !client.Enabled() {
		t.Fatal("expected client to be enabled")
	}
	if got := client.VideoDetailKey(7); got != "myfeed:video:detail:7" {
		t.Fatalf("default-prefixed key = %q", got)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
