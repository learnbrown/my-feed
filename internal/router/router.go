// 只进行路由装配
package router

import (
	"my_feed/internal/account"
	"my_feed/internal/comment"
	"my_feed/internal/feed"
	"my_feed/internal/follow"
	"my_feed/internal/like"
	"my_feed/internal/message"
	"my_feed/internal/middleware"
	"my_feed/internal/profile"
	"my_feed/internal/video"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetRouter(sqlDB *gorm.DB) (router *gin.Engine) {
	router = gin.Default()

	router.StaticFile("/", ".run/template/index.html")

	router.Static("/assets", ".run/assets")
	router.Static("/static/uploads", ".run/uploads")

	router.NoRoute(func(c *gin.Context) {
		c.File("./dist/index.html")
	})

	accountRouter := router.Group("/account")
	accountRepo := account.NewAccountRepo(sqlDB)
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
	videoRepo := video.NewVideoRepo(sqlDB)
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
	likeRepo := like.NewLikeRepo(sqlDB)
	likeService := like.NewLikeService(likeRepo)
	likeHandler := like.NewLikeHandler(likeService)
	{
		likeRouter.POST("/like", likeHandler.Like)
		likeRouter.POST("/unlike", likeHandler.Unlike)
		likeRouter.POST("/isLiked", likeHandler.IsLiked)
		likeRouter.POST("/listLikedVideos", likeHandler.ListLikedVideos)
	}

	commentRouter := router.Group("/comment")
	commentRepo := comment.NewCommentRepo(sqlDB)
	commentService := comment.NewCommentService(commentRepo)
	commentHandler := comment.NewCommentHandler(commentService)
	{
		commentRouter.POST("/listComment", commentHandler.ListComment)

		protected := commentRouter.Group("")
		protected.Use(middleware.JWTAuth(accountRepo))
		{
			protected.POST("/publish", commentHandler.Publish)
			protected.POST("/delete", commentHandler.Delete)
		}
	}

	followRouter := router.Group("/follow")
	followRepo := follow.NewFollowRepo(sqlDB)
	followService := follow.NewFollowService(followRepo, accountRepo)
	followHandler := follow.NewFollowHandler(followService)
	{
		followRouter.POST("/listFollower", followHandler.ListFollower)
		followRouter.POST("/listFollowing", followHandler.ListFollowing)

		protected := followRouter.Group("")
		protected.Use(middleware.JWTAuth(accountRepo))
		{
			protected.POST("/isFollowing", followHandler.IsFollowing)
			protected.POST("/follow", followHandler.Follow)
			protected.POST("/unfollow", followHandler.Unfollow)
		}
	}

	messageRouter := router.Group("/message")
	messageRouter.Use(middleware.JWTAuth(accountRepo))
	messageRepo := message.NewMessageRepo(sqlDB)
	messageService := message.NewMessageService(messageRepo, accountRepo)
	messageHandler := message.NewMessageHandler(messageService)
	{
		messageRouter.POST("/sendMsg", messageHandler.SendMessage)
		messageRouter.POST("/listConversation", messageHandler.ListConversation)
	}

	profileService := profile.NewProfileService(accountRepo, videoRepo, followRepo)
	profileHandler := profile.NewProfileHandler(profileService)
	accountRouter.POST("/getProfile", profileHandler.GetProfile)

	return router
}
