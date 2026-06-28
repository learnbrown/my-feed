package db

import (
	"fmt"
	"my_feed/internal/account"
	"my_feed/internal/comment"
	"my_feed/internal/config"
	"my_feed/internal/follow"
	"my_feed/internal/like"
	"my_feed/internal/message"
	"my_feed/internal/video"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDB(dbcfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbcfg.User, dbcfg.Password, dbcfg.Host, dbcfg.Port, dbcfg.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, err
}

func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&account.Account{},
		&video.Video{},
		&video.Tag{},
		&video.VideoTag{},
		&like.Like{},
		&follow.Follow{},
		&comment.Comment{},
		&message.Message{},
	)
	return err
}

func CloseDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
