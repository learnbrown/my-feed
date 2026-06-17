package router

import (
	"my_feed/internal/account"
	"my_feed/internal/auth"
	"my_feed/internal/middleware"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func InitRouter(db *gorm.DB) (router *gin.Engine) {
	router = gin.Default()

	accountRouter := router.Group("/account")
	accountHandler := NewAccountHandler(db)
	{
		accountRouter.POST("/register", accountHandler.Register)
		accountRouter.POST("/login", accountHandler.Login)

		// 以下路由需鉴权
		protected := accountRouter.Group("")
		protected.Use(middleware.JWTAuth(db))
		{
			protected.POST("/logout", accountHandler.Logout)

			protected.GET("/me", accountHandler.Me)
		}

	}

	return router
}

// todo 怎么在handler中使用数据库
// 定义Handler结构体，将依赖放入
type AccountHandler struct {
	db *gorm.DB
}

func NewAccountHandler(db *gorm.DB) *AccountHandler {
	return &AccountHandler{db: db}
}

// 用户注册提交的信息
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (this *AccountHandler) Register(c *gin.Context) {
	// 从用户处获取注册信息
	var input RegisterRequest
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 去除用户名中的空格
	input.Username = strings.TrimSpace(input.Username)

	// 查询用户名是否存在
	var count int64
	err = this.db.Model(&account.Account{}).
		Where("username = ?", input.Username).
		Count(&count).
		Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if count != 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Username exists"})
		return
	}

	// 创建用户记录
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	acc := account.Account{
		Username:     input.Username,
		PasswordHash: string(passwordHash),
	}

	err = account.CreateAccount(this.db, &acc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

func (this *AccountHandler) Login(c *gin.Context) {
	// 获取用户登录信息
	var input LoginRequest
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Bind input",
			"error":  err.Error(),
		})
		return
	}

	// 查询用户是否存在
	input.Username = strings.TrimSpace(input.Username)

	acc, err := account.FindAccountByName(this.db, input.Username)
	if err != nil {
		// 用户不存在或密码错误返回相同的信息，避免泄漏用户名是否存在的信息
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "Authorization",
			"error":  "Invalid username or password",
		})
		return
	}

	// 比对登录信息
	err = bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(input.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "Authorization",
			"error":  "Invalid username or password",
		})
		return
	}

	// 生成JWTtoken
	token, err := auth.GenerateToken(acc.ID, acc.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Generate token",
			"error":  err.Error()},
		)
		return
	}

	// 更新用户记录
	err = account.UpdateToken(this.db, acc.ID, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Update token",
			"error":  err.Error(),
		})
		return
	}

	// 返回token及用户信息
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"account": gin.H{
			"id":         acc.ID,
			"username":   acc.Username,
			"avatar_url": acc.AvatarURL,
			"bio":        acc.Bio,
		},
	})
}

func (this *AccountHandler) Logout(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Missing authorization context",
		})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "userIDValue->userID error",
			"userIDValue": userIDValue,
		})
		return
	}

	err := account.UpdateToken(this.db, userID, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// /account/me
// 不是让用户传id，而是从上下文取
func (this *AccountHandler) Me(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Missing authorization context",
		})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "userIDValue->userID error",
			"userIDValue": userIDValue,
		})
		return
	}

	acc, err := account.FindAccountByID(this.db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         acc.ID,
		"username":   acc.Username,
		"avatar_url": acc.AvatarURL,
		"bio":        acc.Bio,
	})
}
