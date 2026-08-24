package account

import (
	"context"
	"errors"
	"testing"
)

func TestRegisterRejectsMissingCredentials(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		wantErr  error
	}{
		{name: "empty username", username: "", password: "secret", wantErr: ErrUsernameRequired},
		{name: "whitespace username", username: "  \t", password: "secret", wantErr: ErrUsernameRequired},
		{name: "empty password", username: "alice", password: "", wantErr: ErrPasswordRequired},
	}

	service := &AccountService{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Register(tt.username, tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Register() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoginRejectsMissingCredentials(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
	}{
		{name: "empty username", password: "secret"},
		{name: "whitespace username", username: "  ", password: "secret"},
		{name: "empty password", username: "alice"},
	}

	service := &AccountService{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Login(context.Background(), tt.username, tt.password)
			if !errors.Is(err, ErrInvalidUsernameOrPassword) {
				t.Fatalf("Login() error = %v, want ErrInvalidUsernameOrPassword", err)
			}
		})
	}
}
