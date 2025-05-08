// File: cmd/service/main.go
// @title        Life Is Hard API
// @version      1.0
// @description  這是 Life Is Hard 的後端 API 文件
// @host         localhost:8080
// @BasePath     /api
// @securityDefinitions.oauth2.password OAuth2Password
// @tokenUrl /api/login_user
package main

import (
	"context"
	"log"
	"os"

	"life-is-hard/internal/db"
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

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("環境變數 DATABASE_URL 未設定")
	}

	if err := db.RollbackAll(dbURL); err != nil {
		log.Fatalf("RollbackAll 失敗: %v", err)
	}
	if err := db.RunMigrations(dbURL); err != nil {
		log.Fatalf("Migration 執行失敗: %v", err)
	}

	pool, err := db.NewPool(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("DB 連線失敗: %v", err)
	}
	defer pool.Close()

	e := echo.New()
	// 註冊自定義 validator
	e.Validator = &CustomValidator{validator: validator.New()}

	e.Debug = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	router.Setup(e, pool)

	// Swagger UI endpoint
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.Logger.Fatal(e.Start(":8080"))
}
