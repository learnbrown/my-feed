package feed

import (
	"errors"
	"testing"
	"time"
)

func TestListLatestRejectsOnlyTime(t *testing.T) {
	service := &FeedService{}

	_, err := service.ListLatest(10, time.UnixMilli(1000), 0)
	if !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("expected ErrInvalidCursor, got %v", err)
	}
}

func TestListLatestRejectsOnlyID(t *testing.T) {
	service := &FeedService{}

	_, err := service.ListLatest(10, time.UnixMilli(1000), 0)
	if !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("expected ErrInvalidCursor, got %v", err)
	}
}

func TestNoTag(t *testing.T) {
	service := &FeedService{}
	_, err := service.ListByTag("", 10, time.UnixMilli(0), 0)
	if !errors.Is(err, ErrTagNameRequired) {
		t.Fatalf("expected ErrTagNameRequired, got %v", err)
	}
}
