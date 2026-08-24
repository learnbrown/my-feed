package video

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"
	"time"
)

type detailCacheStub struct {
	detail VideoDTO
	hit    bool
	err    error
}

func (stub *detailCacheStub) GetDetail(context.Context, uint) (VideoDTO, bool, error) {
	return stub.detail, stub.hit, stub.err
}

func (stub *detailCacheStub) SetDetail(context.Context, uint, VideoDTO, time.Duration) error {
	return nil
}

func (stub *detailCacheStub) DelDetail(context.Context, uint) error {
	return nil
}

func TestPublishValidation(t *testing.T) {
	tests := []struct {
		name     string
		authorID uint
		title    string
		playURL  string
		coverURL string
		wantErr  error
	}{
		{name: "author required", title: "title", playURL: "/static/uploads/videos/a.mp4", wantErr: ErrAuthorRequired},
		{name: "title required", authorID: 1, title: "  ", playURL: "/static/uploads/videos/a.mp4", wantErr: ErrTitleRequired},
		{name: "play url required", authorID: 1, title: "title", wantErr: ErrPlayURLRequired},
		{name: "invalid play url", authorID: 1, title: "title", playURL: "https://example.com/a.mp4", wantErr: ErrInvalidPlayURL},
		{name: "invalid cover url", authorID: 1, title: "title", playURL: "/static/uploads/videos/a.mp4", coverURL: "https://example.com/a.jpg", wantErr: ErrInvalidCoverURL},
	}

	service := &VideoService{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Publish(tt.authorID, tt.title, "description", tt.playURL, tt.coverURL)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Publish() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestListByAuthorIDValidation(t *testing.T) {
	service := &VideoService{}

	tests := []struct {
		name       string
		authorID   uint
		latestTime time.Time
		latestID   uint
		wantErr    error
	}{
		{name: "author required", wantErr: ErrAuthorRequired},
		{name: "only time", authorID: 1, latestTime: time.UnixMilli(1000), wantErr: ErrInvalidCursor},
		{name: "only id", authorID: 1, latestID: 1, wantErr: ErrInvalidCursor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ListByAuthorID(tt.authorID, 10, tt.latestTime, tt.latestID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ListByAuthorID() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUploadValidation(t *testing.T) {
	service := &VideoService{}

	tests := []struct {
		name     string
		isCover  bool
		authorID uint
		file     *multipart.FileHeader
		wantErr  error
	}{
		{name: "author required", file: &multipart.FileHeader{Filename: "a.mp4"}, wantErr: ErrAuthorRequired},
		{name: "video file required", authorID: 1, wantErr: ErrFileRequired},
		{name: "video extension", authorID: 1, file: &multipart.FileHeader{Filename: "a.mov"}, wantErr: ErrUnsupportedFileType},
		{name: "cover extension", isCover: true, authorID: 1, file: &multipart.FileHeader{Filename: "a.gif"}, wantErr: ErrUnsupportedFileType},
		{name: "video too large", authorID: 1, file: &multipart.FileHeader{Filename: "a.mp4", Size: maxVideoSize + 1}, wantErr: ErrFileTooLarge},
		{name: "cover too large", isCover: true, authorID: 1, file: &multipart.FileHeader{Filename: "a.png", Size: maxCoverSize + 1}, wantErr: ErrFileTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.isCover {
				_, err = service.UploadCover(tt.authorID, tt.file)
			} else {
				_, err = service.UploadVideo(tt.authorID, tt.file)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("upload error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestToDTO(t *testing.T) {
	createdAt := time.UnixMilli(1_700_000_000_123)
	v := &Video{
		AuthorID:      2,
		Title:         "title",
		Description:   "description",
		PlayURL:       "/static/uploads/videos/a.mp4",
		CoverURL:      "/static/uploads/covers/a.jpg",
		LikesCount:    3,
		CommentsCount: 4,
		CreatedAt:     createdAt,
	}
	v.ID = 1

	dto := ToDTO(v)
	if dto.ID != 1 || dto.AuthorID != 2 || dto.CreatedAt != createdAt.UnixMilli() || dto.LikesCount != 3 || dto.CommentsCount != 4 {
		t.Fatalf("unexpected DTO: %#v", dto)
	}
}

func TestGetDetailReturnsCacheHitWithoutRepository(t *testing.T) {
	want := VideoDTO{ID: 9, Title: "cached"}
	service := NewVideoService(nil, &detailCacheStub{detail: want, hit: true})

	got, err := service.GetDetail(t.Context(), want.ID)
	if err != nil {
		t.Fatalf("GetDetail() error = %v", err)
	}
	if *got != want {
		t.Fatalf("GetDetail() = %#v, want %#v", got, want)
	}
}
