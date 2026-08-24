package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndParseToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "unit-test-secret")

	token, err := GenerateToken(42, "alice")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.AccountID != 42 || claims.Username != "alice" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
	if claims.ExpiresAt == nil || claims.IssuedAt == nil {
		t.Fatalf("expected issued_at and expires_at claims")
	}
	if claims.ID == "" {
		t.Fatal("expected token id claim")
	}
}

func TestGenerateTokenProducesUniqueTokens(t *testing.T) {
	t.Setenv("JWT_SECRET", "unit-test-secret")

	first, err := GenerateToken(42, "alice")
	if err != nil {
		t.Fatalf("first GenerateToken() error = %v", err)
	}
	second, err := GenerateToken(42, "alice")
	if err != nil {
		t.Fatalf("second GenerateToken() error = %v", err)
	}
	if first == second {
		t.Fatal("consecutive logins must not receive the same token")
	}
}

func TestGeneratedFallbackSecretIsStableWithinProcess(t *testing.T) {
	t.Setenv("JWT_SECRET", "")

	token, err := GenerateToken(7, "fallback-user")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("token signed with generated fallback secret should parse: %v", err)
	}
	if claims.AccountID != 7 {
		t.Fatalf("expected account id 7, got %d", claims.AccountID)
	}
}

func TestParseTokenRejectsTamperedToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "unit-test-secret")

	token, err := GenerateToken(42, "alice")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[2] == "" {
		t.Fatalf("unexpected generated token format: %q", token)
	}
	replacement := "A"
	if parts[2][0] == 'A' {
		replacement = "B"
	}
	parts[2] = replacement + parts[2][1:]
	tampered := strings.Join(parts, ".")
	if _, err := ParseToken(tampered); err == nil {
		t.Fatal("expected tampered token to be rejected")
	}
}

func TestParseTokenRejectsUnexpectedSigningMethod(t *testing.T) {
	t.Setenv("JWT_SECRET", "unit-test-secret")

	claims := Claims{
		AccountID: 1,
		Username:  "alice",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString([]byte("unit-test-secret"))
	if err != nil {
		t.Fatalf("sign HS512 token: %v", err)
	}

	if _, err := ParseToken(token); err == nil {
		t.Fatal("expected token with unexpected signing method to be rejected")
	}
}

func TestParseTokenRejectsExpiredToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "unit-test-secret")

	claims := Claims{
		AccountID: 1,
		Username:  "alice",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("unit-test-secret"))
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}

	if _, err := ParseToken(token); err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}
