// 只进行路由装配
package router

import (
	"my_feed/internal/account"
	"my_feed/internal/feed"
	"my_feed/internal/like"
	"my_feed/internal/middleware"
	"my_feed/internal/video"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitRouter(db *gorm.DB) (router *gin.Engine) {
	router = gin.Default()

	router.StaticFile("/", ".run/template/index.html")

	router.Static("/assets", ".run/assets")
	router.Static("/static/uploads", ".run/uploads")

	router.NoRoute(func(c *gin.Context) {
		c.File("./dist/index.html")
	})

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

	likeRouter := router.Group("/like")
	likeRouter.Use(middleware.JWTAuth(accountRepo))
	likeRepo := like.NewLikeRepo(db)
	likeService := like.NewLikeService(likeRepo)
	likeHandler := like.NewLikeHandler(likeService)
	{
		likeRouter.POST("/like", likeHandler.Like)
		likeRouter.POST("/unlike", likeHandler.Unlike)
		likeRouter.POST("/isLiked", likeHandler.IsLiked)
		likeRouter.POST("/listLikedVideos", likeHandler.ListLikedVideos)
	}

	return router
}
