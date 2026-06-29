package comment

import (
	"errors"
	"strings"
	"testing"
)

func TestNoContent(t *testing.T) {
	service := CommentService{}
	_, _, err := service.CreateComment(1, 1, "")
	if !errors.Is(err, ErrContentRequired) {
		t.Fatalf("expected ErrContentRequired, got %v", err)
	}
}

func TestLargeContent(t *testing.T) {
	service := CommentService{}
	content := strings.Repeat("a", 501)
	_, _, err := service.CreateComment(1, 1, content)
	if !errors.Is(err, ErrContentTooLarge) {
		t.Fatalf("expected ErrContentTooLarge, got %v", err)
	}
}
