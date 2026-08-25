package jwt_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"my_feed/internal/account"
	"my_feed/internal/auth"
	"my_feed/internal/cache"
	"my_feed/internal/config"
	jwtmiddleware "my_feed/internal/middleware/jwt"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
)

type tokenCacheStub struct {
	token    string
	hit      bool
	getErr   error
	setErr   error
	getCalls int
	setCalls int
	setToken string
	setTTL   time.Duration
}

func (stub *tokenCacheStub) GetToken(context.Context, uint) (string, bool, error) {
	stub.getCalls++
	return stub.token, stub.hit, stub.getErr
}

func (stub *tokenCacheStub) SetToken(_ context.Context, _ uint, token string, ttl time.Duration) error {
	stub.setCalls++
	stub.setToken = token
	stub.setTTL = ttl
	return stub.setErr
}

func (stub *tokenCacheStub) DelToken(context.Context, uint) error {
	return nil
}

var _ account.TokenCache = (*tokenCacheStub)(nil)

type accountFinderStub struct {
	acc   *account.Account
	err   error
	calls int
}

func (stub *accountFinderStub) FindAccountByID(uint) (*account.Account, error) {
	stub.calls++
	return stub.acc, stub.err
}

func TestJWTAuthRejectsMissingOrMalformedHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		header string
	}{
		{name: "missing"},
		{name: "missing bearer", header: "token"},
		{name: "wrong scheme", header: "Basic token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := protectedRouter(nil, nil)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			router.ServeHTTP(recorder, req)
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401; body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestJWTAuthUsesMatchingCachedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "middleware-test-secret")

	token, _, err := auth.GenerateToken(12, "alice")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	finder := &accountFinderStub{acc: &account.Account{Token: token}}
	router := protectedRouter(finder, &tokenCacheStub{token: token, hit: true})
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body=%s", recorder.Code, recorder.Body.String())
	}
	if finder.calls != 0 {
		t.Fatalf("database lookup calls = %d, want 0 on matching cache hit", finder.calls)
	}
}

func TestJWTAuthFallsBackToMySQLOnCacheMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "middleware-test-secret")

	token, _, err := auth.GenerateToken(12, "alice")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	cacheStub := &tokenCacheStub{token: "another-token", hit: true}
	finder := &accountFinderStub{acc: &account.Account{Token: token}}
	router := protectedRouter(finder, cacheStub)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body=%s", recorder.Code, recorder.Body.String())
	}
	if finder.calls != 1 {
		t.Fatalf("database lookup calls = %d, want 1", finder.calls)
	}
	if cacheStub.setCalls != 1 || cacheStub.setToken != token {
		t.Fatalf("cache refill = (%d, %q), want current token once", cacheStub.setCalls, cacheStub.setToken)
	}
	if cacheStub.setTTL <= 0 || cacheStub.setTTL > account.MaxTokenCacheTTL {
		t.Fatalf("cache refill TTL = %s, want (0, %s]", cacheStub.setTTL, account.MaxTokenCacheTTL)
	}
}

func TestJWTAuthRejectsTokenWhenCacheAndMySQLMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "middleware-test-secret")

	token, _, err := auth.GenerateToken(12, "alice")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	cacheStub := &tokenCacheStub{token: "cached-old-token", hit: true}
	finder := &accountFinderStub{acc: &account.Account{Token: "database-other-token"}}
	router := protectedRouter(finder, cacheStub)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body=%s", recorder.Code, recorder.Body.String())
	}
	if finder.calls != 1 || cacheStub.setCalls != 0 {
		t.Fatalf("database calls = %d, cache set calls = %d; want 1 and 0", finder.calls, cacheStub.setCalls)
	}
}

func TestJWTAuthFallsBackToMySQLOnCacheMissOrError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "middleware-test-secret")

	tests := []struct {
		name  string
		cache *tokenCacheStub
	}{
		{name: "miss", cache: &tokenCacheStub{}},
		{name: "read error", cache: &tokenCacheStub{getErr: errors.New("redis unavailable")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, _, err := auth.GenerateToken(12, "alice")
			if err != nil {
				t.Fatalf("GenerateToken() error = %v", err)
			}
			finder := &accountFinderStub{acc: &account.Account{Token: token}}
			router := protectedRouter(finder, tt.cache)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			router.ServeHTTP(recorder, req)
			if recorder.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want 204; body=%s", recorder.Code, recorder.Body.String())
			}
			if finder.calls != 1 || tt.cache.setCalls != 1 {
				t.Fatalf("database calls = %d, cache set calls = %d; want 1 and 1", finder.calls, tt.cache.setCalls)
			}
		})
	}
}

func TestJWTAuthRejectsStaleCachedTokenAfterShortTTLExpires(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "middleware-test-secret")

	token, _, err := auth.GenerateToken(12, "alice")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	mini := miniredis.RunT(t)
	host, portString, err := net.SplitHostPort(mini.Addr())
	if err != nil {
		t.Fatalf("split miniredis address: %v", err)
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		t.Fatalf("parse miniredis port: %v", err)
	}
	redisClient, err := cache.NewRedis(&config.RedisConfig{
		Host: host, Port: port, KeyPrefix: "test", Enabled: true,
	})
	if err != nil {
		t.Fatalf("NewRedis() error = %v", err)
	}
	t.Cleanup(func() { _ = redisClient.Close() })
	tokenCache := account.NewRedisTokenCache(redisClient)
	if err := tokenCache.SetToken(t.Context(), 12, token, time.Minute); err != nil {
		t.Fatalf("seed token cache: %v", err)
	}

	finder := &accountFinderStub{acc: &account.Account{Token: ""}}
	router := protectedRouter(finder, tokenCache)
	first := httptest.NewRecorder()
	firstRequest := httptest.NewRequest(http.MethodGet, "/protected", nil)
	firstRequest.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(first, firstRequest)
	if first.Code != http.StatusNoContent {
		t.Fatalf("status while stale cache exists = %d, want 204", first.Code)
	}
	if finder.calls != 0 {
		t.Fatalf("database lookup calls while cache exists = %d, want 0", finder.calls)
	}

	mini.FastForward(time.Minute + time.Second)
	second := httptest.NewRecorder()
	secondRequest := httptest.NewRequest(http.MethodGet, "/protected", nil)
	secondRequest.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(second, secondRequest)
	if second.Code != http.StatusUnauthorized {
		t.Fatalf("status after stale cache expires = %d, want 401; body=%s", second.Code, second.Body.String())
	}
	if finder.calls != 1 {
		t.Fatalf("database lookup calls after cache expiry = %d, want 1", finder.calls)
	}
}

func protectedRouter(accountFinder jwtmiddleware.AccountFinder, tokenCache account.TokenCache) *gin.Engine {
	router := gin.New()
	router.GET("/protected", jwtmiddleware.JWTAuth(accountFinder, tokenCache), func(c *gin.Context) {
		if c.GetUint("userID") != 12 || c.GetString("username") != "alice" {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})
	return router
}
