package db

import (
	//"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB() (db *gorm.DB) {
	// todo 将数据库配置放入配置文件

	//dsn := "dev_user:qwerdf@(127.0.0.1:3306)/db001?charset=utf8mb4&parseTime=True&loc=Local"

	//db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	db, err := gorm.Open(sqlite.Open("./data.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return db
}
