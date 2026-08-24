package profile

import (
	"errors"
	"testing"
)

func TestGetProfileRequiresAccount(t *testing.T) {
	service := ProfileService{}

	if _, err := service.GetProfile(0); !errors.Is(err, ErrAccountRequired) {
		t.Fatalf("GetProfile() error = %v, want ErrAccountRequired", err)
	}
}
