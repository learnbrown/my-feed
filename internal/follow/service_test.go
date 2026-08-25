package follow

import (
	"errors"
	"testing"
	"time"
)

func TestSelfFollow(t *testing.T) {
	service := FollowService{}
	err := service.Follow(t.Context(), 1, 1)
	if !errors.Is(err, ErrSelfFollowing) {
		t.Fatalf("expected ErrSelfFollowing, got %v", err)
	}
}

func TestIsFollowingMyself(t *testing.T) {
	service := FollowService{}
	isFollowing, err := service.IsFollowing(1, 1)
	if err != nil {
		t.Fatalf("IsFollowing() error = %v", err)
	}
	if isFollowing {
		t.Fatalf("expected false, got %t", isFollowing)
	}
}

func TestFollowIDValidation(t *testing.T) {
	service := FollowService{}

	tests := []struct {
		name       string
		followerID uint
		vloggerID  uint
		wantErr    error
	}{
		{name: "follower required", vloggerID: 1, wantErr: ErrFollowerRequired},
		{name: "vlogger required", followerID: 1, wantErr: ErrVloggerRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := service.Follow(t.Context(), tt.followerID, tt.vloggerID); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Follow() error = %v, want %v", err, tt.wantErr)
			}
			if err := service.Unfollow(t.Context(), tt.followerID, tt.vloggerID); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Unfollow() error = %v, want %v", err, tt.wantErr)
			}
			if _, err := service.IsFollowing(tt.followerID, tt.vloggerID); !errors.Is(err, tt.wantErr) {
				t.Fatalf("IsFollowing() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestListFollowRequiresAccountID(t *testing.T) {
	service := FollowService{}

	if _, err := service.ListFollower(0, 10, time.Time{}, 0); !errors.Is(err, ErrVloggerRequired) {
		t.Fatalf("ListFollower() error = %v, want ErrVloggerRequired", err)
	}
	if _, err := service.ListFollowing(0, 10, time.Time{}, 0); !errors.Is(err, ErrFollowerRequired) {
		t.Fatalf("ListFollowing() error = %v, want ErrFollowerRequired", err)
	}
}
