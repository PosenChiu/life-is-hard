// File: internal/router/router.go
package router

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"life-is-hard/internal/handler"
	"life-is-hard/internal/handler/auth"
	"life-is-hard/internal/handler/users"
	"life-is-hard/internal/middleware"
)

// Setup 註冊所有路由與中介層
func Setup(e *echo.Echo, db *pgxpool.Pool, rdb *redis.Client) {
	api := e.Group("/api")

	// 健康檢查（需登入）
	api.GET("/ping", handler.PingHandler(db, rdb), middleware.RequireAuth)

	// 使用者登入
	api.POST("/auth/login", auth.AuthLoginHandler(db))

	// 管理員專屬 Users CRUD
	api.POST("/users", users.CreateUserHandler(db), middleware.RequireAdmin)
	api.GET("/users/:id", users.GetUserHandler(db), middleware.RequireAdmin)
	api.PUT("/users/:id", users.UpdateUserHandler(db), middleware.RequireAdmin)
	api.DELETE("/users/:id", users.DeleteUserHandler(db), middleware.RequireAdmin)
	api.POST("/users/:id/reset_password", users.ResetUserPasswordHandler(db), middleware.RequireAdmin)

	// 取得、更新、刪除當前使用者個人資料
	api.GET("/users/me", users.GetMeHandler(db), middleware.RequireAuth)
	api.PUT("/users/me", users.UpdateMeHandler(db), middleware.RequireAuth)
	api.DELETE("/users/me", users.DeleteMeHandler(db), middleware.RequireAuth)
	api.PATCH("/users/me/password", users.UpdatePasswordMeHandler(db), middleware.RequireAuth)
}
