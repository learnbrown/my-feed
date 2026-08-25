package profile

import (
	"errors"
	"net"
	"reflect"
	"strconv"
	"testing"
	"time"

	"my_feed/internal/cache"
	"my_feed/internal/config"

	"github.com/alicebob/miniredis/v2"
)

func TestRedisProfileCacheDisabled(t *testing.T) {
	profileCache := NewRedisProfileCache(nil)

	if _, _, err := profileCache.GetProfile(t.Context(), 1); !errors.Is(err, cache.ErrDisabled) {
		t.Fatalf("GetProfile() error = %v, want ErrDisabled", err)
	}
	if err := profileCache.SetProfile(t.Context(), 1, Profile{}, time.Minute); !errors.Is(err, cache.ErrDisabled) {
		t.Fatalf("SetProfile() error = %v, want ErrDisabled", err)
	}
	if err := profileCache.DelProfile(t.Context(), 1); !errors.Is(err, cache.ErrDisabled) {
		t.Fatalf("DelProfile() error = %v, want ErrDisabled", err)
	}
}

func TestRedisProfileCacheLifecycleTTLAndCorruptValueCleanup(t *testing.T) {
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

	profileCache := NewRedisProfileCache(client)
	want := Profile{
		Account: &ProfileAccount{ID: 9, Username: "cached", AvatarURL: "/avatar.png", Bio: "bio"},
		Stats:   &ProfileStats{VideosCount: 2, LikesCount: 3, FollowersCount: 4, FollowingsCount: 5},
	}

	if got, hit, err := profileCache.GetProfile(t.Context(), want.Account.ID); err != nil || hit || !reflect.DeepEqual(got, Profile{}) {
		t.Fatalf("initial GetProfile() = (%#v, %t, %v), want miss", got, hit, err)
	}
	if err := profileCache.SetProfile(t.Context(), want.Account.ID, want, time.Minute); err != nil {
		t.Fatalf("SetProfile() error = %v", err)
	}
	if got, hit, err := profileCache.GetProfile(t.Context(), want.Account.ID); err != nil || !hit || !reflect.DeepEqual(got, want) {
		t.Fatalf("GetProfile() = (%#v, %t, %v), want %#v", got, hit, err, want)
	}

	key := client.ProfileKey(want.Account.ID)
	if ttl := mini.TTL(key); ttl != time.Minute {
		t.Fatalf("profile TTL = %s, want 1m", ttl)
	}
	mini.FastForward(10 * time.Second)
	if _, hit, err := profileCache.GetProfile(t.Context(), want.Account.ID); err != nil || !hit {
		t.Fatalf("GetProfile() after FastForward hit = %t, error = %v", hit, err)
	}
	if ttl := mini.TTL(key); ttl != 50*time.Second {
		t.Fatalf("profile TTL after hit = %s, want 50s without renewal", ttl)
	}

	if err := profileCache.DelProfile(t.Context(), want.Account.ID); err != nil {
		t.Fatalf("DelProfile() error = %v", err)
	}
	if mini.Exists(key) {
		t.Fatal("profile cache was not deleted")
	}

	if err := client.SetBytes(t.Context(), key, []byte("not-json"), time.Minute); err != nil {
		t.Fatalf("write corrupt cache value: %v", err)
	}
	if got, hit, err := profileCache.GetProfile(t.Context(), want.Account.ID); err == nil || hit || !reflect.DeepEqual(got, Profile{}) {
		t.Fatalf("corrupt GetProfile() = (%#v, %t, %v), want decode error and miss", got, hit, err)
	}
	if mini.Exists(key) {
		t.Fatal("corrupt profile cache was not deleted")
	}
}
