package like

import (
	"errors"
	"testing"
	"time"
)

func TestLikeAndUnlikeValidateIDs(t *testing.T) {
	service := LikeService{}

	tests := []struct {
		name      string
		accountID uint
		videoID   uint
		wantErr   error
	}{
		{name: "author required", videoID: 1, wantErr: ErrAuthorRequired},
		{name: "video required", accountID: 1, wantErr: ErrVideoRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := service.Like(t.Context(), tt.accountID, tt.videoID); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Like() error = %v, want %v", err, tt.wantErr)
			}
			if _, err := service.Unlike(t.Context(), tt.accountID, tt.videoID); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Unlike() error = %v, want %v", err, tt.wantErr)
			}
			if _, err := service.IsLiked(tt.accountID, tt.videoID); !errors.Is(err, tt.wantErr) {
				t.Fatalf("IsLiked() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestListLikedVideosValidation(t *testing.T) {
	service := LikeService{}

	tests := []struct {
		name       string
		accountID  uint
		latestTime time.Time
		latestID   uint
		wantErr    error
	}{
		{name: "author required", wantErr: ErrAuthorRequired},
		{name: "only time", accountID: 1, latestTime: time.UnixMilli(1000), wantErr: ErrInvalidCursor},
		{name: "only id", accountID: 1, latestID: 1, wantErr: ErrInvalidCursor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ListLikedVideos(tt.accountID, 10, tt.latestTime, tt.latestID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ListLikedVideos() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestLikedListToDTOs(t *testing.T) {
	createdAt := time.UnixMilli(1_700_000_000_100)
	likedAt := time.UnixMilli(1_700_000_000_200)
	items := []LikedList{{LikeID: 9, LikedAt: likedAt}}
	items[0].ID = 3
	items[0].AuthorID = 4
	items[0].CreatedAt = createdAt

	dtos := ToDTOs(items)
	if len(dtos) != 1 || dtos[0].ID != 3 || dtos[0].AuthorID != 4 || dtos[0].CreatedAt != createdAt.UnixMilli() || dtos[0].LikedAt != likedAt.UnixMilli() {
		t.Fatalf("unexpected DTOs: %#v", dtos)
	}
}
