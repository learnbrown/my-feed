package comment

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	service *CommentService
}

func NewCommentHandler(service *CommentService) *CommentHandler {
	return &CommentHandler{service: service}
}

// 发布评论请求类型
type PublishRequest struct {
	VideoID uint   `json:"video_id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func (handler *CommentHandler) Publish(c *gin.Context) {
	accountID := c.GetUint("userID")
	if accountID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user id doesn't exist in context",
		})
		return
	}

	input := &PublishRequest{}
	err := c.ShouldBindJSON(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	comment, cnt, err := handler.service.CreateComment(accountID, input.VideoID, input.Content)
	if err != nil {
		switch {
		case errors.Is(err, ErrAuthorRequired):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrVideoRequired) ||
			errors.Is(err, ErrCommentRequired) ||
			errors.Is(err, ErrContentRequired) ||
			errors.Is(err, ErrContentTooLarge):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrVideoNotFound):
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

	c.JSON(http.StatusCreated, gin.H{
		"comment":        comment,
		"comments_count": cnt,
	})
}

// 删除评论请求格式
type DeleteRequest struct {
	CommentID uint `json:"comment_id" binding:"required"`
}

func (handler *CommentHandler) Delete(c *gin.Context) {
	accountID := c.GetUint("userID")
	if accountID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user id doesn't exist in context",
		})
		return
	}

	input := &DeleteRequest{}
	err := c.ShouldBindJSON(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	cnt, err := handler.service.DeleteComment(accountID, input.CommentID)
	if err != nil {
		switch {
		case errors.Is(err, ErrAuthorRequired):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrVideoRequired) ||
			errors.Is(err, ErrCommentRequired) ||
			errors.Is(err, ErrContentTooLarge):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrVideoNotFound) ||
			errors.Is(err, ErrCommentNotFound):
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
		"status":         "ok",
		"comments_count": cnt,
	})
}

// listComment请求格式
type ListCommentRequest struct {
	VideoID    uint  `json:"video_id" binding:"required"`
	Limit      int   `json:"limit"`
	LatestTime int64 `json:"latest_time"`
}

func (handler *CommentHandler) ListComment(c *gin.Context) {
	input := &ListCommentRequest{}
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

	res, err := handler.service.ListComment(input.VideoID, input.Limit, latestTime)
	if err != nil {
		switch {
		case errors.Is(err, ErrVideoRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comments":  res.Comments,
		"has_more":  res.HasMore,
		"next_time": res.NextTime,
	})
}
