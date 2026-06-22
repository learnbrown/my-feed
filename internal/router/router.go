// 只进行路由装配
package router

import (
	"my_feed/internal/account"
	"my_feed/internal/feed"
	"my_feed/internal/middleware"
	"my_feed/internal/video"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitRouter(db *gorm.DB) (router *gin.Engine) {
	router = gin.Default()

	// // 允许所有来源、所有方法的跨域请求
	// config := cors.DefaultConfig()
	// config.AllowAllOrigins = true
	// config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept"}
	// router.Use(cors.New(config))

	// router.Delims("[[", "]]")
	// router.LoadHTMLFiles(".run/template/index.html")

	router.Static("/static/uploads", ".run/uploads")

	// router.GET("/", func(c *gin.Context) {
	// 	c.HTML(200, "index.html", gin.H{})
	// })

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
			protected.POST("/uploadVideo", videoHandler.UploadVideo)
			protected.POST("/uploadCover", videoHandler.UploadCover)
		}
	}

	feedRouter := router.Group("/feed")
	feedService := feed.NewFeedService(videoRepo)
	feedHandler := feed.NewFeedHandler(feedService)
	{
		feedRouter.POST("/listLatest", feedHandler.ListLatest)
		feedRouter.POST("/listByTag", feedHandler.ListByTag)
	}

	return router
}
