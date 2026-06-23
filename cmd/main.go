package main

import (
	"my_feed/internal/account"
	"my_feed/internal/comment"
	"my_feed/internal/db"
	"my_feed/internal/follow"
	"my_feed/internal/like"
	"my_feed/internal/message"
	"my_feed/internal/router"
	"my_feed/internal/video"
)

func main() {
	db := db.InitDB()
	db.AutoMigrate(
		&account.Account{},
		&video.Video{},
		&video.Tag{},
		&video.VideoTag{},
		&like.Like{},
		&follow.Follow{},
		&comment.Comment{},
		&message.Message{},
	)

	r := router.InitRouter(db)

	r.Run()
}
