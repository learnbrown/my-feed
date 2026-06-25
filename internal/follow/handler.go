package follow

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type FollowHandler struct {
	service *FollowService
}

func NewFollowHandler(service *FollowService) *FollowHandler {
	return &FollowHandler{service: service}
}

// isFollowing/follow/unfollow 请求格式
type FollowRequest struct {
	VloggerID uint `json:"vlogger_id" binding:"required"`
}

func (handler *FollowHandler) IsFollowing(c *gin.Context) {
	followerID := c.GetUint("userID")
	if followerID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user id doesn't exist in context",
		})
		return
	}

	input := &FollowRequest{}
	err := c.ShouldBindJSON(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	following, err := handler.service.IsFollowing(followerID, input.VloggerID)
	if err != nil {
		switch {
		case errors.Is(err, ErrFollowerRequired):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrVloggerRequired):
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
		"is_following": following,
	})
}

func (handler *FollowHandler) Follow(c *gin.Context) {
	followerID := c.GetUint("userID")
	if followerID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user id doesn't exist in context",
		})
		return
	}

	input := &FollowRequest{}
	err := c.ShouldBindJSON(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = handler.service.Follow(followerID, input.VloggerID)
	if err != nil {
		switch {
		case errors.Is(err, ErrFollowerRequired):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrVloggerRequired):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrVloggerNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrSelfFollowing):
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "ok",
	})
}

func (handler *FollowHandler) Unfollow(c *gin.Context) {
	followerID := c.GetUint("userID")
	if followerID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user id doesn't exist in context",
		})
		return
	}

	input := &FollowRequest{}
	err := c.ShouldBindJSON(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = handler.service.Unfollow(followerID, input.VloggerID)
	if err != nil {
		switch {
		case errors.Is(err, ErrFollowerRequired):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrVloggerRequired):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrSelfFollowing):
			c.JSON(http.StatusUnprocessableEntity, gin.H{
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
		"status": "ok",
	})
}

// listFollower/listFollowing 请求格式
type ListFollowRequest struct {
	AccountID  uint  `json:"account_id" binding:"required"`
	Limit      int   `json:"limit"`
	LatestTime int64 `json:"latest_time"`
}

// 查看某用户的粉丝列表
func (handler *FollowHandler) ListFollower(c *gin.Context) {
	input := &ListFollowRequest{}
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

	res, err := handler.service.ListFollower(input.AccountID, input.Limit, latestTime)
	if err != nil {
		switch {
		case errors.Is(err, ErrVloggerRequired):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrAccountNotFound):
			c.JSON(http.StatusNotFound, gin.H{
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
		"accounts":  res.Follows,
		"has_more":  res.HasMore,
		"next_time": res.NextTime,
	})
}

// 查看某用户的关注列表
func (handler *FollowHandler) ListFollowing(c *gin.Context) {
	input := &ListFollowRequest{}
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

	res, err := handler.service.ListFollowing(input.AccountID, input.Limit, latestTime)
	if err != nil {
		switch {
		case errors.Is(err, ErrFollowerRequired):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrAccountNotFound):
			c.JSON(http.StatusNotFound, gin.H{
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
		"accounts":  res.Follows,
		"has_more":  res.HasMore,
		"next_time": res.NextTime,
	})
}
