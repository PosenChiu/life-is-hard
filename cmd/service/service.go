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
	"fmt"
	"log"
	"os"
	"strconv"

	"life-is-hard/internal/cache"
	"life-is-hard/internal/database"
	"life-is-hard/internal/router"
	"life-is-hard/internal/worker"

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

var (
	newPgxPool      = database.NewPgxPool
	newRedisClient  = cache.NewRedisClient
	runMigrationsFn = database.RunMigrations
	startServer     = func(e *echo.Echo, addr string) error { return e.Start(addr) }
	newWorkerPool   = worker.NewPool
	exitFunc        = os.Exit
)

func run() error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("環境變數 DATABASE_URL 未設定")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		return fmt.Errorf("環境變數 REDIS_ADDR 未設定")
	}

	redisDBStr := os.Getenv("REDIS_DB")
	if redisDBStr == "" {
		return fmt.Errorf("環境變數 REDIS_DB 未設定")
	}
	redisIndex, err := strconv.Atoi(redisDBStr)
	if err != nil {
		return fmt.Errorf("無效的 REDIS_DB: %v", err)
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		return fmt.Errorf("環境變數 REDIS_PASSWORD 未設定")
	}

	workerCount := 1
	if v := os.Getenv("WORKER_COUNT"); v != "" {
		c, err := strconv.Atoi(v)
		if err != nil || c <= 0 {
			return fmt.Errorf("無效的 WORKER_COUNT: %v", err)
		}
		workerCount = c
	}

	db, err := newPgxPool(context.Background(), dbURL)
	if err != nil {
		return fmt.Errorf("DB 連線失敗: %v", err)
	}
	defer db.Close()

	redis, err := newRedisClient(redisAddr, redisPassword, redisIndex)
	if err != nil {
		return fmt.Errorf("Redis 連線失敗: %v", err)
	}
	defer redis.Close()

	if err := runMigrationsFn(dbURL); err != nil {
		return fmt.Errorf("Migration 執行失敗: %v", err)
	}

	wp := newWorkerPool(workerCount)
	defer wp.Stop()

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.Debug = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	router.Setup(e, db, redis)

	e.GET("/swagger/*", echoSwagger.WrapHandler)
	return startServer(e, ":8080")
}

func main() {
	if err := run(); err != nil {
		log.Print(err)
		exitFunc(1)
	}
}
