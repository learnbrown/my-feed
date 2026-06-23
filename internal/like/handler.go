package like

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type LikeHandler struct {
	service *LikeService
}

func NewLikeHandler(service *LikeService) *LikeHandler {
	return &LikeHandler{service: service}
}

// 点赞请求
type LikeRequest struct {
	VideoID uint `json:"video_id" binding:"required"`
}

func (handler *LikeHandler) Like(c *gin.Context) {
	accountID := c.GetUint("userID")
	if accountID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user id doesn't exist in context",
		})
		return
	}

	input := &LikeRequest{}
	err := c.ShouldBindJSON(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	likesCount, err := handler.service.Like(accountID, input.VideoID)
	if err != nil {
		switch {
		case errors.Is(err, ErrVideoNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrAuthorRequired):
			c.JSON(http.StatusUnauthorized, gin.H{
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
		"status":      "ok",
		"likes_count": likesCount,
	})
}

func (handler *LikeHandler) Unlike(c *gin.Context) {
	accountID := c.GetUint("userID")
	if accountID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user id doesn't exist in context",
		})
		return
	}

	input := &LikeRequest{}
	err := c.ShouldBindJSON(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	likesCount, err := handler.service.Unlike(accountID, input.VideoID)
	if err != nil {
		switch {
		case errors.Is(err, ErrVideoNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrAuthorRequired):
			c.JSON(http.StatusUnauthorized, gin.H{
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
		"status":      "ok",
		"likes_count": likesCount,
	})
}

func (handler *LikeHandler) IsLiked(c *gin.Context) {
	accountID := c.GetUint("userID")
	if accountID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user id doesn't exist in context",
		})
		return
	}

	input := &LikeRequest{}
	err := c.ShouldBindJSON(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	isLiked, err := handler.service.IsLiked(accountID, input.VideoID)
	if err != nil {
		switch {
		case errors.Is(err, ErrAuthorRequired):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrVideoRequired):
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
		"is_liked": isLiked,
	})
}

// ListLiked请求体格式
type ListLikedRequest struct {
	Limit      int   `json:"limit"`
	LatestTime int64 `json:"latest_time"`
}

func (handler *LikeHandler) ListLikedVideos(c *gin.Context) {
	accountID := c.GetUint("userID")
	if accountID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user id doesn't exist in context",
		})
		return
	}

	input := &ListLikedRequest{}
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

	res, err := handler.service.ListLikedVideos(accountID, input.Limit, latestTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"likes":     res.Likes,
		"has_more":  res.HasMore,
		"next_time": res.NextTime,
	})
}
