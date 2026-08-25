package profile

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	service *ProfileService
}

func NewProfileHandler(service *ProfileService) *ProfileHandler {
	return &ProfileHandler{service: service}
}

// getProfile请求格式
type ProfileRequest struct {
	AccountID uint `json:"account_id" binding:"required"`
}

func (handler *ProfileHandler) GetProfile(c *gin.Context) {
	input := &ProfileRequest{}
	err := c.ShouldBindJSON(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	profile, err := handler.service.GetProfile(c.Request.Context(), input.AccountID)
	if err != nil {
		switch {
		case errors.Is(err, ErrAccountRequired):
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

	c.JSON(http.StatusOK, profile)
}
