package cache

import "testing"

func TestKeysUseConfiguredPrefix(t *testing.T) {
	client := &Client{keyPrefix: "testfeed"}

	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "account token", got: client.AccountTokenKey(12), want: "testfeed:account:token:12"},
		{name: "video detail", got: client.VideoDetailKey(34), want: "testfeed:video:detail:34"},
		{name: "profile", got: client.ProfileKey(56), want: "testfeed:profile:56"},
		{name: "latest feed", got: client.FeedLatestKey(), want: "testfeed:feed:latest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestNilClientKeysDoNotPanic(t *testing.T) {
	var client *Client

	if got := client.AccountTokenKey(1); got != "account:token:1" {
		t.Fatalf("AccountTokenKey() = %q", got)
	}
	if got := client.VideoDetailKey(2); got != "video:detail:2" {
		t.Fatalf("VideoDetailKey() = %q", got)
	}
	if got := client.ProfileKey(3); got != "profile:3" {
		t.Fatalf("ProfileKey() = %q", got)
	}
	if got := client.FeedLatestKey(); got != "feed:latest" {
		t.Fatalf("FeedLatestKey() = %q", got)
	}
}

func TestDisabledClientOperationsReturnErrDisabled(t *testing.T) {
	var client *Client

	if _, err := client.GetBytes(t.Context(), "key"); err != ErrDisabled {
		t.Fatalf("GetBytes() error = %v, want ErrDisabled", err)
	}
	if err := client.SetBytes(t.Context(), "key", []byte("value"), 0); err != ErrDisabled {
		t.Fatalf("SetBytes() error = %v, want ErrDisabled", err)
	}
	if err := client.Del(t.Context(), "key"); err != ErrDisabled {
		t.Fatalf("Del() error = %v, want ErrDisabled", err)
	}
	if err := client.Ping(t.Context()); err != ErrDisabled {
		t.Fatalf("Ping() error = %v, want ErrDisabled", err)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
