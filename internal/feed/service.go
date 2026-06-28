package feed

import (
	"errors"
	"my_feed/internal/video"
	"strings"
	"time"
)

type FeedService struct {
	repo *video.VideoRepo
}

func NewFeedService(repo *video.VideoRepo) *FeedService {
	return &FeedService{repo: repo}
}

var (
	ErrTagNameRequired = errors.New("tag name required")
	ErrInvalidCursor   = errors.New("invalid cursor")
)

// [x]: 升级为created_at + id 复合游标
func (service *FeedService) ListLatest(limit int, latestTime time.Time, latestID uint) (*video.ListResponse, error) {
	if limit > 50 {
		limit = 50
	}
	if limit <= 0 {
		limit = 20
	}

	if latestTime.IsZero() != (latestID == 0) {
		return nil, ErrInvalidCursor
	}

	videos, err := service.repo.ListLatest(limit+1, latestTime, latestID)
	if err != nil {
		return nil, err
	}

	hasMore := len(*videos) > limit
	if hasMore {
		*videos = (*videos)[:limit]
	}

	var nextTime int64
	var nextID uint
	if len(*videos) > 0 {
		nextTime = (*videos)[len(*videos)-1].CreatedAt.UnixMilli()
		nextID = (*videos)[len(*videos)-1].ID
	}

	dtos := video.ToDTOs(*videos)

	res := &video.ListResponse{
		Videos:   dtos,
		NextTime: nextTime,
		NextID:   nextID,
		HasMore:  hasMore,
	}

	return res, nil
}

// 实现 /feed/listByTag：
// 请求：tag_name, limit, latest_time
// repo 查询路径：tags -> video_tags -> videos
// 排序和分页逻辑复用 listLatest 的思路。
// 查不到 tag 时返回空列表，不要返回 500。

// [x]: 升级为created_at + id 复合游标
func (service *FeedService) ListByTag(tagName string, limit int, latestTime time.Time, latestID uint) (*video.ListResponse, error) {
	tagName = strings.TrimSpace(tagName)
	// 兼容输入 `#GO`
	tagName = strings.TrimPrefix(tagName, "#")
	tagName = strings.ToLower(tagName)
	if tagName == "" {
		return nil, ErrTagNameRequired
	}

	if limit > 50 {
		limit = 50
	}
	if limit <= 0 {
		limit = 20
	}

	if latestTime.IsZero() != (latestID == 0) {
		return nil, ErrInvalidCursor
	}

	// 调用repo ListByTag
	videos, err := service.repo.ListByTag(tagName, limit+1, latestTime, latestID)
	if err != nil {
		return nil, err
	}

	hasMore := len(*videos) > limit
	if hasMore {
		*videos = (*videos)[:limit]
	}

	var nextTime int64
	var nextID uint
	if len(*videos) > 0 {
		nextTime = (*videos)[len(*videos)-1].CreatedAt.UnixMilli()
		nextID = (*videos)[len(*videos)-1].ID
	}

	dtos := video.ToDTOs(*videos)

	res := &video.ListResponse{
		Videos:   dtos,
		NextTime: nextTime,
		NextID:   nextID,
		HasMore:  hasMore,
	}

	return res, nil
}
