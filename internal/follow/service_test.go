package follow

import (
	"errors"
	"testing"
)

func TestSelfFollow(t *testing.T) {
	service := FollowService{}
	err := service.Follow(1, 1)
	if !errors.Is(err, ErrSelfFollowing) {
		t.Fatalf("expected ErrSelfFollowing, got %v", err)
	}
}

func TestIsFollowingMyself(t *testing.T) {
	service := FollowService{}
	isFollowing, _ := service.IsFollowing(1, 1)
	if isFollowing {
		t.Fatalf("expected false, got %t", isFollowing)
	}
}
