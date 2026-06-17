package main

import (
	"my_feed/internal/account"
	"my_feed/internal/db"
	"my_feed/internal/router"
)

func main() {
	db := db.InitDB()
	db.AutoMigrate(&account.Account{})

	r := router.InitRouter(db)

	r.Run()
}