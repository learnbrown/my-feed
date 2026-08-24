package feed

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestListHandlersRejectNegativeCursorAndStop(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		path string
		body string
	}{
		{name: "latest", path: "/feed/listLatest", body: `{"latest_time":-1,"latest_id":0}`},
		{name: "by tag", path: "/feed/listByTag", body: `{"tag_name":"go","latest_time":-1,"latest_id":0}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &FeedService{}
			handler := NewFeedHandler(service)
			router := gin.New()
			router.POST("/feed/listLatest", handler.ListLatest)
			router.POST("/feed/listByTag", handler.ListByTag)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(recorder, req)

			assertSingleInvalidCursorResponse(t, recorder)
		})
	}
}

func assertSingleInvalidCursorResponse(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", recorder.Code, recorder.Body.String())
	}
	decoder := json.NewDecoder(strings.NewReader(recorder.Body.String()))
	var body map[string]string
	if err := decoder.Decode(&body); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, recorder.Body.String())
	}
	if body["error"] != ErrInvalidCursor.Error() {
		t.Fatalf("error = %q, want %q", body["error"], ErrInvalidCursor.Error())
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		t.Fatalf("expected exactly one JSON response, got trailing data: %v", err)
	}
}
