package message

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	service *MessageService
}

func NewMessageHandler(service *MessageService) *MessageHandler {
	return &MessageHandler{service: service}
}

// SendMessage 请求格式
type SendMsgRequest struct {
	ToID    uint   `json:"to_id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func (handler *MessageHandler) SendMessage(c *gin.Context) {
	fromID := c.GetUint("userID")
	if fromID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user id doesn't exist in context",
		})
		return
	}

	input := &SendMsgRequest{}
	err := c.ShouldBindJSON(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	msg, err := handler.service.SendMessage(fromID, input.ToID, input.Content)
	if err != nil {
		switch {
		case errors.Is(err, ErrSenderRequired):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrReceiverRequired) ||
			errors.Is(err, ErrContentRequired) ||
			errors.Is(err, ErrContentTooLarge):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrAccountNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrSendToYourself):
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
		"message": msg,
	})
}

// listConversation 请求格式
type ListCvsRequest struct {
	ToID       uint  `json:"to_id" binding:"required"`
	Limit      int   `json:"limit"`
	LatestTime int64 `json:"latest_time"`
	LatestID   uint  `json:"latest_id"`
}

func (handler *MessageHandler) ListConversation(c *gin.Context) {
	fromID := c.GetUint("userID")
	if fromID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user id doesn't exist in context",
		})
		return
	}

	input := &ListCvsRequest{}
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

	res, err := handler.service.ListConversation(fromID, input.ToID, input.Limit, latestTime, input.LatestID)
	if err != nil {
		switch {
		case errors.Is(err, ErrSenderRequired):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrReceiverRequired) ||
			errors.Is(err, ErrInvalidCursor):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrAccountNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrSendToYourself):
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

	c.JSON(http.StatusOK, res)
}
