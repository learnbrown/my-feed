package video

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestListByAuthorIDRejectsNegativeCursorAndStops(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &VideoService{}
	handler := NewVideoHandler(service)
	router := gin.New()
	router.POST("/video/listByAuthorID", handler.ListByAuthorID)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/video/listByAuthorID", strings.NewReader(`{"author_id":1,"latest_time":-1,"latest_id":0}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)

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

func TestUploadHandlersSaveFilesUnderConfiguredRoot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		path          string
		formField     string
		filename      string
		responseField string
		content       string
	}{
		{name: "video", path: "/video/uploadVideo", formField: "video", filename: "sample.mp4", responseField: "play_url", content: "video-bytes"},
		{name: "cover", path: "/video/uploadCover", formField: "cover", filename: "sample.png", responseField: "cover_url", content: "image-bytes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uploadRoot := t.TempDir()
			service := &VideoService{uploadRoot: uploadRoot}
			handler := NewVideoHandler(service)
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("userID", uint(42))
				c.Next()
			})
			router.POST("/video/uploadVideo", handler.UploadVideo)
			router.POST("/video/uploadCover", handler.UploadCover)

			req := newMultipartUploadRequest(t, tt.path, tt.formField, tt.filename, tt.content)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body=%s", recorder.Code, recorder.Body.String())
			}
			var response map[string]string
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			url := response[tt.responseField]
			if !strings.HasPrefix(url, "/static/uploads/") {
				t.Fatalf("unexpected upload URL %q", url)
			}
			relativePath := strings.TrimPrefix(url, "/static/uploads/")
			stored, err := os.ReadFile(filepath.Join(uploadRoot, filepath.FromSlash(relativePath)))
			if err != nil {
				t.Fatalf("read uploaded file: %v", err)
			}
			if string(stored) != tt.content {
				t.Fatalf("stored content = %q, want %q", stored, tt.content)
			}
		})
	}
}

func newMultipartUploadRequest(t *testing.T, path, field, filename, content string) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(field, filename)
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write multipart content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}
