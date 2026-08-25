package integration

import (
	"errors"
	"testing"
	"time"

	dbtest "my_feed/internal/db/testutil"
	"my_feed/internal/video"
)

func TestVideoDetailCacheAsideWithDB(t *testing.T) {
	sqlDB := dbtest.SetupTestDB(t)
	dbtest.CleanTestDB(t, sqlDB)

	author := createTestAccount(t, sqlDB, "cache-aside-author")
	v := createTestVideo(t, sqlDB, author.ID, 1, 1, fixedCursorTime())
	repo := video.NewVideoRepo(sqlDB)

	t.Run("miss reads database and fills cache", func(t *testing.T) {
		cacheRecorder := &detailCacheRecorder{}
		service := video.NewVideoService(repo, cacheRecorder, nil)

		got, err := service.GetDetail(t.Context(), v.ID)
		if err != nil {
			t.Fatalf("GetDetail() error = %v", err)
		}
		if got.ID != v.ID || got.Title != v.Title {
			t.Fatalf("unexpected detail: %#v", got)
		}
		if cacheRecorder.setVideoID != v.ID || cacheRecorder.setDetail.ID != v.ID || cacheRecorder.setTTL != 5*time.Minute {
			t.Fatalf("unexpected cache fill: id=%d detail=%#v ttl=%s", cacheRecorder.setVideoID, cacheRecorder.setDetail, cacheRecorder.setTTL)
		}
	})

	t.Run("cache failure degrades to database", func(t *testing.T) {
		cacheRecorder := &detailCacheRecorder{getErr: errors.New("redis unavailable")}
		service := video.NewVideoService(repo, cacheRecorder, nil)

		got, err := service.GetDetail(t.Context(), v.ID)
		if err != nil {
			t.Fatalf("GetDetail() error = %v", err)
		}
		if got.ID != v.ID || cacheRecorder.setVideoID != v.ID {
			t.Fatalf("expected database fallback and cache refill, got=%#v cacheID=%d", got, cacheRecorder.setVideoID)
		}
	})

	t.Run("hit bypasses database result", func(t *testing.T) {
		cached := video.VideoDTO{ID: v.ID, Title: "cached-title"}
		cacheRecorder := &detailCacheRecorder{getDetail: cached, getHit: true}
		service := video.NewVideoService(repo, cacheRecorder, nil)

		got, err := service.GetDetail(t.Context(), v.ID)
		if err != nil {
			t.Fatalf("GetDetail() error = %v", err)
		}
		if *got != cached {
			t.Fatalf("GetDetail() = %#v, want cached %#v", got, cached)
		}
		if cacheRecorder.setVideoID != 0 {
			t.Fatalf("cache hit unexpectedly wrote cache for video %d", cacheRecorder.setVideoID)
		}
	})
}
