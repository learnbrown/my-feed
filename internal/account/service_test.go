package account

import (
	"errors"
	"testing"
)

func TestRegisterRequiresPassword(t *testing.T) {
	service := &AccountService{}

	_, err := service.Register("user1", "")
	if !errors.Is(err, ErrPasswordRequired) {
		t.Fatalf("expected ErrPasswordRequired, got %v", err)
	}
}
