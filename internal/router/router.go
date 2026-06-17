package router

import (
	"my_feed/internal/account"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// todo 怎么在handler中使用数据库
func InitRouter(db *gorm.DB) (router *gin.Engine) {
	router = gin.Default()

	
	accountRouter := router.Group("/account")
	{
		accountRouter.POST("/register", RegisterHandler)
		accountRouter.POST("/login", LoginHandler)
		accountRouter.POST("/logout", LogoutHandler)

		accountRouter.GET("/me", MeHandler)

	}

	return router
}

func RegisterHandler(c *gin.Context) {
	var acc account.Account
	err := c.ShouldBind(&acc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	account.CreateAccount()

	c.JSON(http.StatusOK, gin.H{"status": "ok"})

}

func LoginHandler(c *gin.Context) {

}

func LogoutHandler(c *gin.Context) {

}

func MeHandler(c *gin.Context) {

}
