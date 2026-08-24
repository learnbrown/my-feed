package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"my_feed/internal/account"
	"my_feed/internal/video"

	"gorm.io/gorm"
)

type detailCacheRecorder struct {
	deletedVideoIDs []uint
	getDetail       video.VideoDTO
	getHit          bool
	getErr          error
	setVideoID      uint
	setDetail       video.VideoDTO
	setTTL          time.Duration
}

func (recorder *detailCacheRecorder) GetDetail(context.Context, uint) (video.VideoDTO, bool, error) {
	return recorder.getDetail, recorder.getHit, recorder.getErr
}

func (recorder *detailCacheRecorder) SetDetail(_ context.Context, videoID uint, detail video.VideoDTO, ttl time.Duration) error {
	recorder.setVideoID = videoID
	recorder.setDetail = detail
	recorder.setTTL = ttl
	return nil
}

func (recorder *detailCacheRecorder) DelDetail(_ context.Context, videoID uint) error {
	recorder.deletedVideoIDs = append(recorder.deletedVideoIDs, videoID)
	return nil
}

func createTestAccount(t *testing.T, sqlDB *gorm.DB, username string) account.Account {
	t.Helper()

	acc := account.Account{
		Username:     username,
		PasswordHash: "test-password-hash",
	}
	if err := sqlDB.Create(&acc).Error; err != nil {
		t.Fatalf("create account %q: %v", username, err)
	}
	return acc
}

func createTestVideo(t *testing.T, sqlDB *gorm.DB, authorID uint, sequence int, status int, createdAt time.Time) video.Video {
	t.Helper()

	v := video.Video{
		AuthorID:  authorID,
		Title:     fmt.Sprintf("video-%d", sequence),
		PlayURL:   fmt.Sprintf("/static/uploads/videos/test/%d.mp4", sequence),
		CoverURL:  "/static/uploads/covers/default.png",
		Status:    status,
		CreatedAt: createdAt,
	}
	if err := sqlDB.Create(&v).Error; err != nil {
		t.Fatalf("create video %d: %v", sequence, err)
	}
	return v
}

func reverseUintIDs(ids []uint) []uint {
	reversed := make([]uint, len(ids))
	for i := range ids {
		reversed[len(ids)-1-i] = ids[i]
	}
	return reversed
}

func assertUintIDs(t *testing.T, got, want []uint) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("ids length = %d, want %d; got=%v want=%v", len(got), len(want), got, want)
	}
	seen := make(map[uint]struct{}, len(got))
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ids = %v, want %v", got, want)
		}
		if _, exists := seen[got[i]]; exists {
			t.Fatalf("duplicate id %d in %v", got[i], got)
		}
		seen[got[i]] = struct{}{}
	}
}

func fixedCursorTime() time.Time {
	return time.Date(2026, time.August, 24, 12, 0, 0, 123_000_000, time.Local)
}
