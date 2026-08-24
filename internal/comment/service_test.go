package comment

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCreateCommentValidation(t *testing.T) {
	tests := []struct {
		name      string
		accountID uint
		videoID   uint
		content   string
		wantErr   error
	}{
		{name: "author required", videoID: 1, content: "hello", wantErr: ErrAuthorRequired},
		{name: "video required", accountID: 1, content: "hello", wantErr: ErrVideoRequired},
		{name: "empty content", accountID: 1, videoID: 1, content: " \t\n", wantErr: ErrContentRequired},
		{name: "content over rune limit", accountID: 1, videoID: 1, content: strings.Repeat("界", 501), wantErr: ErrContentTooLarge},
	}

	service := CommentService{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := service.CreateComment(t.Context(), tt.accountID, tt.videoID, tt.content)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateComment() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteCommentValidation(t *testing.T) {
	service := CommentService{}

	if _, err := service.DeleteComment(t.Context(), 0, 1); !errors.Is(err, ErrAuthorRequired) {
		t.Fatalf("DeleteComment() error = %v, want ErrAuthorRequired", err)
	}
	if _, err := service.DeleteComment(t.Context(), 1, 0); !errors.Is(err, ErrCommentRequired) {
		t.Fatalf("DeleteComment() error = %v, want ErrCommentRequired", err)
	}
}

func TestListCommentValidation(t *testing.T) {
	service := CommentService{}

	tests := []struct {
		name       string
		videoID    uint
		latestTime time.Time
		latestID   uint
		wantErr    error
	}{
		{name: "video required", wantErr: ErrVideoRequired},
		{name: "only time", videoID: 1, latestTime: time.UnixMilli(1000), wantErr: ErrInvalidCursor},
		{name: "only id", videoID: 1, latestID: 1, wantErr: ErrInvalidCursor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ListComment(tt.videoID, 10, tt.latestTime, tt.latestID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ListComment() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
