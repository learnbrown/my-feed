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
			errors.Is(err, ErrPlayURLRequired):

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

func (handler *VideoHandler) GetDetail(c *gin.Context) {
	input := &DetailRequest{}
	err := c.ShouldBindJSON(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	video, err := handler.service.GetDetail(input.ID)
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
// TODO: LatestTime是做什么的
// 游标分页，见README
type ListRequest struct {
	AuthorID   uint  `json:"author_id" binding:"required"`
	Limit      int   `json:"limit" binding:"required"`
	LatestTime int64 `json:"latest_time"`
}

type ListResponse struct {
	Videos   []Video `json:"videos"`
	NextTime int64   `json:"next_time"`
	HasMore  bool    `json:"has_more"`
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

	var latestTime time.Time
	if input.LatestTime > 0 {
		latestTime = time.UnixMilli(input.LatestTime)
	}

	res, err := handler.service.ListByAuthorID(input.AuthorID, input.Limit, latestTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	// [x] 为什么列表为空
	c.JSON(http.StatusOK, gin.H{
		"videos":    res.Videos,
		"next_time": res.NextTime,
		"has_more":  res.HasMore,
	})
}

// Upload video
func (handler *VideoHandler) UploadVideo(c *gin.Context) {
 
}

// Upload cover
func (handler *VideoHandler) UploadCover(c *gin.Context) {

}
