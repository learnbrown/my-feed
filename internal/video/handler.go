package video

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
	service *VideoService
}

func NewVideoHandler(service *VideoService) *VideoHandler {
	return &VideoHandler{service: service}
}

// Publish video
type PublishRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	PlayURL     string `json:"play_url" binding:"required"`
	CoverURL    string `json:"cover_url"`
}

func (handler *VideoHandler) PublishVideo(c *gin.Context) {
	authorID := c.GetUint("userID")
	if authorID == 0 {
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

	video, err := handler.service.Publish(authorID, input.Title, input.Description, input.PlayURL, input.CoverURL)
	if err != nil {
		switch {
		case errors.Is(err, ErrAuthorRequired):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrTitleRequired) ||
			errors.Is(err, ErrPlayURLRequired) ||
			errors.Is(err, ErrInvalidPlayURL) ||
			errors.Is(err, ErrInvalidCoverURL):

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

	c.JSON(http.StatusCreated, gin.H{
		"video": video,
	})
}

// Get video detail
type DetailRequest struct {
	ID uint `json:"id" binding:"required"`
}

// TODO: 更常见的用法为 GET /video/getDetail/:id，先用POST
func (handler *VideoHandler) GetDetail(c *gin.Context) {
	input := &DetailRequest{}
	err := c.ShouldBindJSON(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	video, err := handler.service.GetDetail(c.Request.Context(), input.ID)
	if err != nil {
		switch {
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
	c.JSON(http.StatusOK, gin.H{
		"video": video,
	})
}

// Get video list of author
// [x] LatestTime是做什么的
// 游标分页，见README
// [x] 升级为 created_at + id 复合游标

type ListRequest struct {
	AuthorID   uint  `json:"author_id" binding:"required"`
	Limit      int   `json:"limit"`
	LatestTime int64 `json:"latest_time"`
	LatestID   uint  `json:"latest_id"`
}

func (handler *VideoHandler) ListByAuthorID(c *gin.Context) {
	input := &ListRequest{}
	err := c.ShouldBindJSON(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if input.LatestTime < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrInvalidCursor.Error(),
		})
		return
	}

	var latestTime time.Time
	if input.LatestTime > 0 {
		latestTime = time.UnixMilli(input.LatestTime)
	}

	res, err := handler.service.ListByAuthorID(input.AuthorID, input.Limit, latestTime, input.LatestID)
	if err != nil {
		switch {
		case errors.Is(err, ErrAuthorRequired) ||
			errors.Is(err, ErrInvalidCursor):
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

	// [x] 为什么列表为空
	c.JSON(http.StatusOK, res)
}

// Upload video
func (handler *VideoHandler) UploadVideo(c *gin.Context) {
	authorID := c.GetUint("userID")
	if authorID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user id doesn't exist in context",
		})
		return
	}
	// [x] 有几种错误还没处理 ErrFileRequired
	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	returnDir, err := handler.service.UploadVideo(authorID, file)
	if err != nil {
		switch {
		case errors.Is(err, ErrAuthorRequired):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrUnsupportedFileType) ||
			errors.Is(err, ErrFileTooLarge) ||
			errors.Is(err, ErrFileRequired):
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
		"play_url": returnDir,
	})
}

// Upload cover
func (handler *VideoHandler) UploadCover(c *gin.Context) {
	authorID := c.GetUint("userID")
	if authorID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user id doesn't exist in context",
		})
		return
	}

	file, err := c.FormFile("cover")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	returnDir, err := handler.service.UploadCover(authorID, file)
	if err != nil {
		switch {
		case errors.Is(err, ErrAuthorRequired):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrUnsupportedFileType) ||
			errors.Is(err, ErrFileTooLarge) ||
			errors.Is(err, ErrFileRequired):
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
		"cover_url": returnDir,
	})
}
