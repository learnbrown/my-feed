package main

import (
	"log"
	"my_feed/internal/config"
	"my_feed/internal/db"
	"my_feed/internal/router"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func main() {
	// 加载.env
	err := godotenv.Load()
	if err != nil {
		log.Println(".env not found, continue")
	}

	// 加载config
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}
	log.Printf("Loading config from %s", configPath)
	cfg, useDefault, err := config.LoadLoaclDev(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if useDefault {
		log.Printf("Config file %s not found, using default local config", configPath)
	} else {
		log.Printf("Config loaded from file: %s", configPath)
	}

	// 连接数据库
	sqlDB, err := db.NewDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	err = db.AutoMigrate(sqlDB)
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}

	// 设置路由
	r := router.SetRouter(sqlDB)

	err = r.Run(":" + strconv.Itoa(cfg.Server.Port))
	if err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
