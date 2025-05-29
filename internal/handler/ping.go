package handler

import (
	"net/http"
	"time"

	"life-is-hard/internal/api"
	"life-is-hard/internal/cache"
	"life-is-hard/internal/database"

	"github.com/labstack/echo/v4"
)

// @Summary     Health Check
// @Description 回傳 pong，並檢查資料庫連線是否正常，同時在 Redis 設置一個示例鍵值
// @Tags        ping
// @Accept      json
// @Produce     json
// @Success     200 {object} api.PingResponse
// @Failure     500 {object} api.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /ping [get]
func PingHandler(db database.DB, cache cache.Cache) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		if err := db.Ping(ctx); err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: "database unhealthy"})
		}

		if err := cache.Set(ctx, "ping:timestamp", time.Now().Format(time.RFC3339), time.Minute).Err(); err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: "cache unhealthy"})
		}

		return c.JSON(http.StatusOK, api.PingResponse{Message: "pong"})
	}
}
