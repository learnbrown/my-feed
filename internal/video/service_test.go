package video_test

import (
	"errors"
	"my_feed/internal/video"
	"testing"
)

func TestNoTitle(t *testing.T) {
	service := &video.VideoService{}
	_, err := service.Publish(1, "", "test", "test", "test")
	if !errors.Is(err, video.ErrTitleRequired) {
		t.Fatalf("expected ErrTitleRequired, got %v", err)
	}
}

func TestNoPlayURL(t *testing.T) {
	service := &video.VideoService{}
	_, err := service.Publish(1, "test", "test", "", "test")
	if !errors.Is(err, video.ErrPlayURLRequired) {
		t.Fatalf("expected ErrPlayURLRequired, got %v", err)
	}
}

func TestInvalidPlayURL(t *testing.T) {
	service := &video.VideoService{}
	_, err := service.Publish(1, "test", "test", "invalid url", "test")
	if !errors.Is(err, video.ErrInvalidPlayURL) {
		t.Fatalf("expected ErrInvalidPlayURL, got %v", err)
	}
}

func TestInvalidCoverURL(t *testing.T) {
	service := &video.VideoService{}
	_, err := service.Publish(1, "test", "test", "/static/uploads/videos/test", "test")
	if !errors.Is(err, video.ErrInvalidCoverURL) {
		t.Fatalf("expected ErrInvalidCoverURL, got %v", err)
	}
}
