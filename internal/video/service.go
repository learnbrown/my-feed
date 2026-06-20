package video

import (
	"errors"
	"my_feed/internal/db"
	"strings"
	"time"
)

type VideoService struct {
	repo *VideoRepo
}

func NewVideoService(repo *VideoRepo) *VideoService {
	return &VideoService{repo: repo}
}

var (
	ErrVideoNotFound   = errors.New("video not found")
	ErrAuthorNotFound  = errors.New("author not found")
	ErrAuthorRequired  = errors.New("author required")
	ErrTitleRequired   = errors.New("title required")
	ErrPlayURLRequired = errors.New("play url required")
)

func (service *VideoService) Publish(authorID uint, title, description, video_url, cover_url string) (*Video, error) {
	// [x] 缺少业务校验
	if authorID == 0 {
		return nil, ErrAuthorRequired
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrTitleRequired
	}
	video_url = strings.TrimSpace(video_url)
	if video_url == "" {
		return nil, ErrPlayURLRequired
	}

	cover_url = strings.TrimSpace(cover_url)
	if cover_url == "" {
		cover_url = "/static/uploads/covers/default.png"
	}

	video := &Video{
		AuthorID:    authorID,
		Title:       title,
		Description: description,
		PlayURL:     video_url,
		CoverURL:    cover_url,
	}

	err := service.repo.CreateVideo(video)
	return video, err
}

func (service *VideoService) GetDetail(id uint) (*Video, error) {
	video, err := service.repo.FindVideoByID(id)
	if errors.Is(err, db.ErrRecordNotFound) {
		return nil, ErrVideoNotFound
	}
	return video, err
}

func (service *VideoService) ListByAuthorID(authorID uint, limit int, latestTime time.Time) (*ListResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	videos, err := service.repo.ListByAuthorID(authorID, limit, latestTime)
	// TODO: find即使没查到数据，也不会返回ErrRecordNotFound，而是返回空列表
	// 先不做处理
	// if errors.Is(err, db.ErrRecordNotFound) {
	// 	return nil, ErrAuthorNotFound
	// }

	var nextTime int64
	if len(*videos) > 0 {
		nextTime = (*videos)[len(*videos)-1].CreatedAt.UnixMilli()
	}

	res := &ListResponse{
		Videos: *videos, 
		NextTime: nextTime, 
		HasMore: len(*videos) == limit,
	}
	
	return res, err
}
