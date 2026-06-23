package feed

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type FeedHandler struct {
	service *FeedService
}

func NewFeedHandler(service *FeedService) *FeedHandler {
	return &FeedHandler{service: service}
}

// list_latest请求格式
type ListLatestRequest struct {
	Limit      int   `json:"limit"`
	LatestTime int64 `json:"latest_time"`
}

// 首页返回最新视频
func (handler *FeedHandler) ListLatest(c *gin.Context) {
	input := &ListLatestRequest{}
	err := c.ShouldBindJSON(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// int64转化为毫秒时间戳
	var latestTime time.Time
	if input.LatestTime > 0 {
		latestTime = time.UnixMilli(input.LatestTime)
	}

	res, err := handler.service.ListLatest(input.Limit, latestTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"videos":    res.Videos,
		"next_time": res.NextTime,
		"has_more":  res.HasMore,
	})
}

// list_by_tag 请求格式
type ListByTagRequest struct {
	TagName    string `json:"tag_name" binding:"required"`
	Limit      int    `json:"limit"`
	LatestTime int64  `json:"latest_time"`
}

func (handler *FeedHandler) ListByTag(c *gin.Context) {
	input := &ListByTagRequest{}
	err := c.ShouldBindJSON(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var latestTime time.Time
	if input.LatestTime > 0 {
		latestTime = time.UnixMilli(input.LatestTime)
	}

	res, err := handler.service.ListByTag(input.TagName, input.Limit, latestTime)
	if err != nil {
		switch {
		case errors.Is(err, ErrTagNameRequired):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		default:

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"videos":    res.Videos,
		"next_time": res.NextTime,
		"has_more":  res.HasMore,
	})
}

// SELECT * FROM `videos` WHERE id IN (SELECT video_id FROM `video_tags` WHERE tag_id IN (SELECT id FROM `tags` WHERE name = "GO")) AND status id = 1 AND `videos`.`deleted_at` IS NULL ORDER BY created_at desc LIMIT 11
