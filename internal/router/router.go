// File: internal/router/router.go
package router

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"life-is-hard/internal/handler"
	"life-is-hard/internal/handler/auth"
	"life-is-hard/internal/handler/oauth"
	"life-is-hard/internal/handler/users"
	"life-is-hard/internal/middleware"
)

// Setup 註冊所有路由與中介層
func Setup(e *echo.Echo, db *pgxpool.Pool, rdb *redis.Client) {
	api := e.Group("/api")

	// 健康檢查（需登入）
	api.GET("/ping", handler.PingHandler(db, rdb), middleware.RequireAuth)

	// 使用者登入
	api.POST("/auth/login", auth.LoginHandler(db))
	api.POST("/oauth/token", oauth.TokenHandler(db, rdb))

	// 管理員專屬 Users CRUD
	apiUsers := api.Group("/users", middleware.RequireAdmin)
	apiUsers.POST("", users.CreateUserHandler(db))
	apiUsers.GET("/:id", users.GetUserHandler(db))
	apiUsers.PUT("/:id", users.UpdateUserHandler(db))
	apiUsers.DELETE("/:id", users.DeleteUserHandler(db))

	// 取得、更新、刪除當前使用者個人資料
	apiUsersMe := api.Group("/users/me", middleware.RequireAuth)
	apiUsersMe.GET("", users.GetMyUserHandler(db))
	apiUsersMe.PUT("", users.UpdateMyUserHandler(db))
	apiUsersMe.DELETE("", users.DeleteMyUserHandler(db))
	apiUsersMe.PATCH("/password", users.UpdateMyUserPasswordHandler(db))

	apiUsersMeClients := apiUsersMe.Group("/oauth/clients", middleware.RequireAuth)
	apiUsersMeClients.POST("", users.CreateMyOAuthClientHandler(db))
	apiUsersMeClients.GET("", users.ListMyOAuthClientsHandler(db))
	apiUsersMeClients.GET("/:client_id", users.GetMyOAuthClientHandler(db))
	apiUsersMeClients.PUT("/:client_id", users.UpdateMyOAuthClientHandler(db))
	apiUsersMeClients.DELETE("/:client_id", users.DeleteMyOAuthClientHandler(db))
}
