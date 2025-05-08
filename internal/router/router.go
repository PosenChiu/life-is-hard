// File: internal/router/router.go
package router

import (
	"life-is-hard/internal/handler"
	"life-is-hard/internal/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// Setup 註冊所有路由與中介層
func Setup(e *echo.Echo, db *pgxpool.Pool, rdb *redis.Client) {
	api := e.Group("/api")

	// 健康檢查（需登入）
	api.GET("/ping", handler.PingHandler(db, rdb), middleware.RequireAuth)

	// 使用者登入
	api.POST("/login_user", handler.LoginUserHandler(db))

	// Users 路由群組（需登入）
	users := api.Group("/users", middleware.RequireAuth)

	// 管理員專屬路由
	users.POST("", handler.CreateUserHandler(db), middleware.RequireAdmin)
	users.GET("/:id", handler.GetUserHandler(db), middleware.RequireAdmin)
	users.PUT("/:id", handler.UpdateUserHandler(db), middleware.RequireAdmin)
	users.DELETE("/:id", handler.DeleteUserHandler(db), middleware.RequireAdmin)

	// 取得、更新、刪除當前使用者個人資料
	users.GET("/me", handler.GetMeHandler(db))
	users.PUT("/me", handler.UpdateMeHandler(db))
	users.DELETE("/me", handler.DeleteMeHandler(db))

	// 更新當前使用者密碼
	users.PATCH("/me/password", handler.UpdatePasswordMeHandler(db))

	// 管理員重置其他使用者密碼
	users.POST("/:id/reset_password", handler.ResetUserPasswordHandler(db), middleware.RequireAdmin)
}
