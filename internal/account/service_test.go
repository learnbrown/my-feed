package account

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type accountRepositoryStub struct {
	account        *Account
	updateTokenErr error
	updatedToken   string
	updateCalls    int
	events         *[]string
}

func (stub *accountRepositoryStub) CreateAccount(*Account) error { return nil }

func (stub *accountRepositoryStub) UpdateToken(_ uint, token string) error {
	stub.updateCalls++
	stub.updatedToken = token
	if stub.events != nil {
		*stub.events = append(*stub.events, "mysql")
	}
	return stub.updateTokenErr
}

func (stub *accountRepositoryStub) FindAccountByID(uint) (*Account, error) {
	return stub.account, nil
}

func (stub *accountRepositoryStub) FindAccountByName(string) (*Account, error) {
	return stub.account, nil
}

func (stub *accountRepositoryStub) ExistAccount(string) (bool, error) { return false, nil }

type tokenCacheServiceStub struct {
	setErr   error
	delErr   error
	setCalls int
	delCalls int
	setTTL   time.Duration
	events   *[]string
}

func (stub *tokenCacheServiceStub) GetToken(context.Context, uint) (string, bool, error) {
	return "", false, nil
}

func (stub *tokenCacheServiceStub) SetToken(_ context.Context, _ uint, _ string, ttl time.Duration) error {
	stub.setCalls++
	stub.setTTL = ttl
	if stub.events != nil {
		*stub.events = append(*stub.events, "redis_set")
	}
	return stub.setErr
}

func (stub *tokenCacheServiceStub) DelToken(context.Context, uint) error {
	stub.delCalls++
	if stub.events != nil {
		*stub.events = append(*stub.events, "redis_delete")
	}
	return stub.delErr
}

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

func TestLoginUpdatesMySQLBeforeRedisAndIgnoresCacheFailure(t *testing.T) {
	t.Setenv("JWT_SECRET", "account-service-test-secret")
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	events := make([]string, 0, 2)
	repo := &accountRepositoryStub{
		account: &Account{Model: gorm.Model{ID: 7}, Username: "alice", PasswordHash: string(passwordHash)},
		events:  &events,
	}
	cacheStub := &tokenCacheServiceStub{setErr: errors.New("redis unavailable"), events: &events}
	service := NewAccountService(repo, cacheStub)

	acc, err := service.Login(context.Background(), "alice", "secret")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if acc.Token == "" || repo.updatedToken != acc.Token {
		t.Fatalf("stored token = %q, returned token = %q", repo.updatedToken, acc.Token)
	}
	if cacheStub.setCalls != 1 || cacheStub.setTTL <= 0 || cacheStub.setTTL > MaxTokenCacheTTL {
		t.Fatalf("cache set calls = %d, TTL = %s", cacheStub.setCalls, cacheStub.setTTL)
	}
	if len(events) != 2 || events[0] != "mysql" || events[1] != "redis_set" {
		t.Fatalf("operation order = %v, want [mysql redis_set]", events)
	}
}

func TestLoginDoesNotWriteCacheWhenMySQLUpdateFails(t *testing.T) {
	t.Setenv("JWT_SECRET", "account-service-test-secret")
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	wantErr := errors.New("mysql unavailable")
	repo := &accountRepositoryStub{
		account:        &Account{Model: gorm.Model{ID: 7}, Username: "alice", PasswordHash: string(passwordHash)},
		updateTokenErr: wantErr,
	}
	cacheStub := &tokenCacheServiceStub{}
	service := NewAccountService(repo, cacheStub)

	if _, err := service.Login(context.Background(), "alice", "secret"); !errors.Is(err, wantErr) {
		t.Fatalf("Login() error = %v, want %v", err, wantErr)
	}
	if cacheStub.setCalls != 0 {
		t.Fatalf("cache set calls = %d, want 0", cacheStub.setCalls)
	}
}

func TestLogoutUpdatesMySQLBeforeRedisAndIgnoresCacheFailure(t *testing.T) {
	events := make([]string, 0, 2)
	repo := &accountRepositoryStub{events: &events}
	cacheStub := &tokenCacheServiceStub{delErr: errors.New("redis unavailable"), events: &events}
	service := NewAccountService(repo, cacheStub)

	if err := service.Logout(context.Background(), 7); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if repo.updatedToken != "" || cacheStub.delCalls != 1 {
		t.Fatalf("updated token = %q, cache delete calls = %d", repo.updatedToken, cacheStub.delCalls)
	}
	if len(events) != 2 || events[0] != "mysql" || events[1] != "redis_delete" {
		t.Fatalf("operation order = %v, want [mysql redis_delete]", events)
	}
}

func TestLogoutDoesNotDeleteCacheWhenMySQLUpdateFails(t *testing.T) {
	wantErr := errors.New("mysql unavailable")
	repo := &accountRepositoryStub{updateTokenErr: wantErr}
	cacheStub := &tokenCacheServiceStub{}
	service := NewAccountService(repo, cacheStub)

	if err := service.Logout(context.Background(), 7); !errors.Is(err, wantErr) {
		t.Fatalf("Logout() error = %v, want %v", err, wantErr)
	}
	if cacheStub.delCalls != 0 {
		t.Fatalf("cache delete calls = %d, want 0", cacheStub.delCalls)
	}
}
