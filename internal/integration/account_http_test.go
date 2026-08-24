package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterRequiresPasswordHTTP(t *testing.T) {
	router := setupTestRouter(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/account/register",
		strings.NewReader(`{"username":"user1"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestAccountAuthLifecycleHTTP(t *testing.T) {
	router := setupTestRouter(t)

	register := performJSONRequest(t, router, http.MethodPost, "/account/register", `{"username":"alice","password":"secret123"}`, "")
	if register.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want 201; body=%s", register.Code, register.Body.String())
	}

	firstLogin := performJSONRequest(t, router, http.MethodPost, "/account/login", `{"username":"alice","password":"secret123"}`, "")
	if firstLogin.Code != http.StatusOK {
		t.Fatalf("first login status = %d, want 200; body=%s", firstLogin.Code, firstLogin.Body.String())
	}
	var firstLoginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(firstLogin.Body.Bytes(), &firstLoginBody); err != nil {
		t.Fatalf("decode first login response: %v", err)
	}
	if firstLoginBody.Token == "" {
		t.Fatal("login response did not contain token")
	}

	me := performJSONRequest(t, router, http.MethodGet, "/account/me", "", firstLoginBody.Token)
	if me.Code != http.StatusOK {
		t.Fatalf("me status = %d, want 200; body=%s", me.Code, me.Body.String())
	}

	secondLogin := performJSONRequest(t, router, http.MethodPost, "/account/login", `{"username":"alice","password":"secret123"}`, "")
	if secondLogin.Code != http.StatusOK {
		t.Fatalf("second login status = %d, want 200; body=%s", secondLogin.Code, secondLogin.Body.String())
	}
	var secondLoginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(secondLogin.Body.Bytes(), &secondLoginBody); err != nil {
		t.Fatalf("decode second login response: %v", err)
	}
	if secondLoginBody.Token == "" || secondLoginBody.Token == firstLoginBody.Token {
		t.Fatal("second login must issue a distinct token")
	}

	meWithOldToken := performJSONRequest(t, router, http.MethodGet, "/account/me", "", firstLoginBody.Token)
	if meWithOldToken.Code != http.StatusUnauthorized {
		t.Fatalf("me with old token status = %d, want 401; body=%s", meWithOldToken.Code, meWithOldToken.Body.String())
	}

	logout := performJSONRequest(t, router, http.MethodPost, "/account/logout", `{}`, secondLoginBody.Token)
	if logout.Code != http.StatusOK {
		t.Fatalf("logout status = %d, want 200; body=%s", logout.Code, logout.Body.String())
	}

	meAfterLogout := performJSONRequest(t, router, http.MethodGet, "/account/me", "", secondLoginBody.Token)
	if meAfterLogout.Code != http.StatusUnauthorized {
		t.Fatalf("me after logout status = %d, want 401; body=%s", meAfterLogout.Code, meAfterLogout.Body.String())
	}
}

func performJSONRequest(t *testing.T, handler http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}
