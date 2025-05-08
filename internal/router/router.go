// File: internal/router/router.go
package router

import (
	"life-is-hard/internal/handler"
	"life-is-hard/internal/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// Setup 註冊所有路由與中介層
func Setup(e *echo.Echo, db *pgxpool.Pool) {
	api := e.Group("/api")
	api.GET("/ping", handler.PingHandler(db), middleware.RequireAuth)
	api.POST("/login_user", handler.LoginUserHandler(db))

	users := api.Group("/users")
	users.POST("", handler.CreateUserHandler(db), middleware.RequireAdmin)
}
