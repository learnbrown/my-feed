package middleware

import (
	"errors"
	"my_feed/internal/account"
	"my_feed/internal/auth"
	"my_feed/internal/db"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuth(accountRepo *account.AccountRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 读取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token needed"})
			// 阻断后续handler
			c.Abort()
			return
		}

		// 检查token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// 解析并验证token
		tokenString := parts[1]
		claims, err := auth.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token or it has expired",
			})
			c.Abort()
			return
		}

		// 检查数据库中token
		acc, err := accountRepo.FindAccountByID(claims.AccountID)
		if err != nil {
			// find返回的错误有多种，需分开处理
			if errors.Is(err, db.ErrRecordNotFound) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "user doesn't exist",
				})
				c.Abort()
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}
		if acc.Token != tokenString {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token or it has expired",
			})
			c.Abort()
			return
		}

		// 将解析得到的用户信息放入gin上下文c中
		c.Set("userID", claims.AccountID)
		c.Set("username", claims.Username)

		// 进入后续handler
		c.Next()
	}
}
