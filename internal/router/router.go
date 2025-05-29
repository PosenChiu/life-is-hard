package router

import (
	"github.com/labstack/echo/v4"

	"life-is-hard/internal/cache"
	"life-is-hard/internal/database"
	"life-is-hard/internal/handler"
	"life-is-hard/internal/handler/auth"
	"life-is-hard/internal/handler/oauth"
	"life-is-hard/internal/handler/users"
	"life-is-hard/internal/middleware"
)

// Setup 註冊所有路由與中介層
func Setup(e *echo.Echo, db database.DB, cache cache.Cache) {
	api := e.Group("/api")

	// 健康檢查（需登入）
	api.GET("/ping", handler.PingHandler(db, cache), middleware.RequireAuth)

	// 使用者登入
	api.POST("/auth/login", auth.LoginHandler(db))
	api.POST("/oauth/token", oauth.TokenHandler(db, cache))

	// 管理員專屬 Users CRUD
	api.POST("/users", users.CreateUserHandler(db), middleware.RequireAdmin)
	api.GET("/users/:id", users.GetUserHandler(db), middleware.RequireAdmin)
	api.PUT("/users/:id", users.UpdateUserHandler(db), middleware.RequireAdmin)
	api.DELETE("/users/:id", users.DeleteUserHandler(db), middleware.RequireAdmin)

	// 取得、更新、刪除當前使用者個人資料
	api.GET("/users/me", users.GetMyUserHandler(db), middleware.RequireAuth)
	api.PUT("/users/me", users.UpdateMyUserHandler(db), middleware.RequireAuth)
	api.DELETE("/users/me", users.DeleteMyUserHandler(db), middleware.RequireAuth)
	api.PATCH("/users/me/password", users.UpdateMyUserPasswordHandler(db), middleware.RequireAuth)

	api.POST("/users/me/oauth-clients", users.CreateMyOAuthClientHandler(db), middleware.RequireAuth)
	api.GET("/users/me/oauth-clients", users.ListMyOAuthClientsHandler(db), middleware.RequireAuth)
	api.GET("/users/me/oauth-clients/:client_id", users.GetMyOAuthClientHandler(db), middleware.RequireAuth)
	api.PUT("/users/me/oauth-clients/:client_id", users.UpdateMyOAuthClientHandler(db), middleware.RequireAuth)
	api.DELETE("/users/me/oauth-clients/:client_id", users.DeleteMyOAuthClientHandler(db), middleware.RequireAuth)
}
