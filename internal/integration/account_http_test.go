package integration

import (
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
