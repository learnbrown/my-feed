// 只进行路由装配
package router

import (
	"my_feed/internal/account"
	"my_feed/internal/middleware"
	"my_feed/internal/video"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitRouter(db *gorm.DB) (router *gin.Engine) {
	router = gin.Default()
	router.Static("/static/uploads", ".run/uploads")

	accountRouter := router.Group("/account")
	accountRepo := account.NewAccountRepo(db)
	accountService := account.NewAccountService(accountRepo)
	accountHandler := account.NewAccountHandler(accountService)
	{
		accountRouter.POST("/register", accountHandler.Register)
		accountRouter.POST("/login", accountHandler.Login)

		// 以下路由需鉴权
		protected := accountRouter.Group("")
		protected.Use(middleware.JWTAuth(accountRepo))
		{
			protected.POST("/logout", accountHandler.Logout)

			protected.GET("/me", accountHandler.Me)
		}

	}

	videoRouter := router.Group("/video")
	videoRepo := video.NewVideoRepo(db)
	videoService := video.NewVideoService(videoRepo)
	videoHandler := video.NewVideoHandler(videoService)
	{
		videoRouter.POST("/getDetail", videoHandler.GetDetail)
		videoRouter.POST("/listByAuthorID", videoHandler.ListByAuthorID)

		protected := videoRouter.Group("")
		protected.Use(middleware.JWTAuth(accountRepo))
		{
			protected.POST("/publish", videoHandler.PublishVideo)

		}
	}

	return router
}
