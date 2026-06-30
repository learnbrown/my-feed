package jwt

import (
	"context"
	"errors"
	"log"
	"my_feed/internal/account"
	"my_feed/internal/auth"
	"my_feed/internal/dberr"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func JWTAuth(accountRepo *account.AccountRepo, tokenCache account.TokenCache) gin.HandlerFunc {
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

		// 检查redis中token
		getCtx, cancel := context.WithTimeout(c.Request.Context(), 50*time.Millisecond)
		defer cancel()

		token, hit, err := tokenCache.GetToken(getCtx, claims.AccountID)
		if err == nil {
			if hit {
				log.Printf("Successfully get token cache")
				if token != tokenString {
					c.JSON(http.StatusUnauthorized, gin.H{
						"error": "Invalid token or it has expired",
					})
					c.Abort()
					return
				}
				c.Set("userID", claims.AccountID)
				c.Set("username", claims.Username)
				c.Next()
				return
			} else {
				log.Printf("Miss token cache")
			}
		} else {
			log.Printf("Failed to get token cache: %v", err)
		}

		// 检查数据库中token
		acc, err := accountRepo.FindAccountByID(claims.AccountID)
		if err != nil {
			// find返回的错误有多种，需分开处理
			if errors.Is(err, dberr.ErrRecordNotFound) {
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

		// 回填redis
		setCtx, cancel := context.WithTimeout(c.Request.Context(), 50*time.Millisecond)
		defer cancel()
		err = tokenCache.SetToken(setCtx, claims.AccountID, tokenString, 2*time.Hour)
		if err != nil {
			log.Printf("Failed to set cache: %v", err)
		} else {
			log.Printf("Successfully set token cache")
		}

		// 将解析得到的用户信息放入gin上下文c中
		c.Set("userID", claims.AccountID)
		c.Set("username", claims.Username)

		// 进入后续handler
		c.Next()
	}
}
