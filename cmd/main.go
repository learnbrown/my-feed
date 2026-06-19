package main

import (
	"my_feed/internal/account"
	"my_feed/internal/db"
	"my_feed/internal/router"
	"my_feed/internal/video"
)

func main() {
	db := db.InitDB()
	db.AutoMigrate(&account.Account{}, &video.Video{}, &video.Tag{}, &video.VideoTag{})

	r := router.InitRouter(db)

	r.Run()
}
