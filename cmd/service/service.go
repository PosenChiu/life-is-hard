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

	"life-is-hard/internal/cache"
	"life-is-hard/internal/database"
	"life-is-hard/internal/router"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

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
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("環境變數 DATABASE_URL 未設定")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("環境變數 REDIS_ADDR 未設定")
	}

	redisDBStr := os.Getenv("REDIS_DB")
	if redisDBStr == "" {
		log.Fatal("環境變數 REDIS_DB 未設定")
	}
	redisIndex, err := strconv.Atoi(redisDBStr)
	if err != nil {
		log.Fatalf("無效的 REDIS_DB: %v", err)
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		log.Fatal("環境變數 REDIS_PASSWORD 未設定")
	}

	db, err := database.NewPgxPool(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("DB 連線失敗: %v", err)
	}
	defer db.Close()

	redis, err := cache.NewRedisClient(redisAddr, redisPassword, redisIndex)
	if err != nil {
		log.Fatalf("Redis 連線失敗: %v", err)
	}
	defer redis.Close()

	if err := database.RunMigrations(dbURL); err != nil {
		log.Fatalf("Migration 執行失敗: %v", err)
	}

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.Debug = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	router.Setup(e, db, redis)

	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.Logger.Fatal(e.Start(":8080"))
}
