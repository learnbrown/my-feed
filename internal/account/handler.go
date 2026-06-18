// HTTP处理
package account

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// todo 怎么在handler中使用数据库
// 定义Handler结构体，将依赖放入
type AccountHandler struct {
	service *AccountService
}

func NewAccountHandler(service *AccountService) *AccountHandler {
	return &AccountHandler{service: service}
}

// 用户注册提交的信息
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (handler *AccountHandler) Register(c *gin.Context) {
	// 从用户处获取注册信息
	var input RegisterRequest
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acc, err := handler.service.Register(input.Username, input.Password)

	if err != nil {
		switch {
		case errors.Is(err, ErrUsernameRequired) || errors.Is(err, ErrPasswordRequired):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, ErrUsernameExists):
			c.JSON(http.StatusConflict, gin.H{
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
		"status":   "ok",
		"id":       acc.ID,
		"username": acc.Username,
	})
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (handler *AccountHandler) Login(c *gin.Context) {
	// 获取用户登录信息
	var input LoginRequest
	err := c.ShouldBindJSON(&input)
	if err != nil {
		// 不应该用500，应该用400表示json错误，缺少字段
		// c.JSON(http.StatusInternalServerError, gin.H{
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bind input",
			"error":  err.Error(),
		})
		return
	}

	acc, err := handler.service.Login(input.Username, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidUsernameOrPassword):
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

	// 返回token及用户信息
	c.JSON(http.StatusOK, gin.H{
		"token": acc.Token,
		"account": gin.H{
			"id":         acc.ID,
			"username":   acc.Username,
			"avatar_url": acc.AvatarURL,
			"bio":        acc.Bio,
		},
	})
}

func (handler *AccountHandler) Logout(c *gin.Context) {
	userID := c.GetUint("userID")

	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user id doesn't exist in context",
		})
		return
	}

	err := handler.service.Logout(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// /account/me
// 不是让用户传id，而是从上下文取
func (handler *AccountHandler) Me(c *gin.Context) {
	userID := c.GetUint("userID")

	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user id doesn't exist in context",
		})
		return
	}

	acc, err := handler.service.Me(userID)
	if err != nil {
		switch {
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
		"id":         acc.ID,
		"username":   acc.Username,
		"avatar_url": acc.AvatarURL,
		"bio":        acc.Bio,
	})
}
