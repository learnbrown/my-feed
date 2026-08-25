package main

import (
	"context"
	"log"
	"my_feed/internal/cache"
	"my_feed/internal/config"
	"my_feed/internal/db"
	"my_feed/internal/router"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// 加载.env
	err := godotenv.Load()
	if err != nil {
		log.Printf("level=INFO component=startup operation=load_env status=not_found action=continue")
	}

	// 加载config
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}
	log.Printf("level=INFO component=startup operation=load_config path=%q", configPath)
	cfg, useDefault, err := config.LoadLoaclDev(configPath)
	if err != nil {
		log.Fatalf("level=FATAL component=startup operation=load_config err=%q", err)
	}
	if useDefault {
		log.Printf("level=INFO component=startup operation=load_config path=%q status=not_found action=use_local_defaults", configPath)
	} else {
		log.Printf("level=INFO component=startup operation=load_config path=%q status=loaded", configPath)
	}

	// 连接数据库
	sqlDB, err := db.NewDB(cfg.Database)
	if err != nil {
		log.Fatalf("level=FATAL component=startup operation=connect_mysql err=%q", err)
	}
	err = db.AutoMigrate(sqlDB)
	if err != nil {
		log.Fatalf("level=FATAL component=startup operation=auto_migrate err=%q", err)
	}
	log.Printf("level=INFO component=startup operation=connect_mysql status=ready")

	// 连接Redis
	var rediscache *cache.Client
	if !cfg.Redis.Enabled {
		log.Printf("level=INFO component=startup operation=connect_redis status=disabled")
	} else {
		rediscache, err = cache.NewRedis(&cfg.Redis)
		if err != nil {
			log.Printf("level=WARN component=startup operation=connect_redis status=disabled err=%q", err)
		} else {
			pingCtx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			pingErr := rediscache.Ping(pingCtx)
			cancel()
			if pingErr != nil {
				log.Printf("level=WARN component=startup operation=ping_redis status=disabled err=%q", pingErr)
				_ = rediscache.Close()
				rediscache = nil
			} else {
				defer rediscache.Close()
				log.Printf("level=INFO component=startup operation=connect_redis status=ready")
			}
		}
	}

	// 设置路由
	r := router.SetRouter(sqlDB, rediscache)

	err = r.Run(":" + strconv.Itoa(cfg.Server.Port))
	if err != nil {
		log.Fatalf("level=FATAL component=startup operation=run_http_server port=%d err=%q", cfg.Server.Port, err)
	}
}
