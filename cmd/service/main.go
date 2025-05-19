// File: cmd/service/main.go
// @title        Life Is Hard API
// @version      1.0
// @description  這是 Life Is Hard 的後端 API 文件
// @host         localhost:8080
// @BasePath     /api
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @securityDefinitions.oauth2.application OAuth2Application
// @tokenUrl /api/oauth/token
// @securityDefinitions.oauth2.password OAuth2Password
// @tokenUrl /api/oauth/token
package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"life-is-hard/internal/db"
	"life-is-hard/internal/router"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"

	_ "life-is-hard/docs" // 引入 swag 產出的 docs

	echoSwagger "github.com/swaggo/echo-swagger"
)

// CustomValidator wraps go-playground/validator for Echo
// swagger:ignore
type CustomValidator struct {
	validator *validator.Validate
}

// Validate calls the underlying validator
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	// 資料庫連線字符串
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("環境變數 DATABASE_URL 未設定")
	}

	// Redis 配置
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("環境變數 REDIS_ADDR 未設定")
	}

	redisDBStr := os.Getenv("REDIS_DB")
	if redisDBStr == "" {
		log.Fatal("環境變數 REDIS_DB 未設定")
	}
	rdbIndex, err := strconv.Atoi(redisDBStr)
	if err != nil {
		log.Fatalf("無效的 REDIS_DB: %v", err)
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		log.Fatal("環境變數 REDIS_PASSWORD 未設定")
	}

	// 回滾並執行遷移
	// if err := db.RollbackAll(dbURL); err != nil {
	// 	log.Fatalf("RollbackAll 失敗: %v", err)
	// }
	if err := db.RunMigrations(dbURL); err != nil {
		log.Fatalf("Migration 執行失敗: %v", err)
	}

	// 建立資料庫連線池
	pool, err := db.NewPool(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("DB 連線失敗: %v", err)
	}
	defer pool.Close()

	// 建立 Redis 客戶端
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       rdbIndex,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis 連線失敗: %v", err)
	}
	defer func() {
		if err := rdb.Close(); err != nil {
			log.Printf("關閉 Redis 連線失敗: %v", err)
		}
	}()

	// Echo 實例及中介層
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.Debug = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 註冊路由並注入 db 與 rdb
	router.Setup(e, pool, rdb)

	// Swagger UI
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// 啟動服務
	e.Logger.Fatal(e.Start(":8080"))
}
