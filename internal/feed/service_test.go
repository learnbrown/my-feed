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

	_, err := service.ListLatest(10, time.Time{}, 1)
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

func TestListByTagRejectsHalfCursor(t *testing.T) {
	tests := []struct {
		name       string
		latestTime time.Time
		latestID   uint
	}{
		{name: "only time", latestTime: time.UnixMilli(1000)},
		{name: "only id", latestID: 1},
	}

	service := &FeedService{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ListByTag("go", 10, tt.latestTime, tt.latestID)
			if !errors.Is(err, ErrInvalidCursor) {
				t.Fatalf("expected ErrInvalidCursor, got %v", err)
			}
		})
	}
}
