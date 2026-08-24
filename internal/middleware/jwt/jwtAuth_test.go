package jwt_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"my_feed/internal/account"
	"my_feed/internal/auth"
	jwtmiddleware "my_feed/internal/middleware/jwt"

	"github.com/gin-gonic/gin"
)

type tokenCacheStub struct {
	token string
	hit   bool
	err   error
}

func (stub *tokenCacheStub) GetToken(context.Context, uint) (string, bool, error) {
	return stub.token, stub.hit, stub.err
}

func (stub *tokenCacheStub) SetToken(context.Context, uint, string, time.Duration) error {
	return nil
}

func (stub *tokenCacheStub) DelToken(context.Context, uint) error {
	return nil
}

var _ account.TokenCache = (*tokenCacheStub)(nil)

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
			router := protectedRouter(nil)
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

	token, err := auth.GenerateToken(12, "alice")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	router := protectedRouter(&tokenCacheStub{token: token, hit: true})
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestJWTAuthRejectsMismatchedCachedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "middleware-test-secret")

	token, err := auth.GenerateToken(12, "alice")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	router := protectedRouter(&tokenCacheStub{token: "another-token", hit: true})
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body=%s", recorder.Code, recorder.Body.String())
	}
}

func protectedRouter(tokenCache account.TokenCache) *gin.Engine {
	router := gin.New()
	router.GET("/protected", jwtmiddleware.JWTAuth(nil, tokenCache), func(c *gin.Context) {
		if c.GetUint("userID") != 12 || c.GetString("username") != "alice" {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})
	return router
}
