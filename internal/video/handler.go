package video

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
	service *VideoService
}

func NewVideoHandler(service *VideoService) *VideoHandler {
	return &VideoHandler{service: service}
}

func (handler *VideoHandler) UploadVideo(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		
	})
}
