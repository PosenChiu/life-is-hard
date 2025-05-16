// File: internal/handler/ping.go
package handler

import (
	"net/http"
	"time"

	"life-is-hard/internal/dto"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// PingHandler 健康檢查（需通過認證），並於 Redis 設定示例鍵值
// @Summary     Health Check
// @Description 回傳 pong，並檢查資料庫連線是否正常，同時在 Redis 設置一個示例鍵值
// @Tags        ping
// @Accept      json
// @Produce     json
// @Success     200 {object} dto.PingResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /ping [get]
func PingHandler(db *pgxpool.Pool, rdb *redis.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		if err := db.Ping(ctx); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: "database unhealthy"})
		}

		err := rdb.Set(ctx, "ping:timestamp", time.Now().Format(time.RFC3339), time.Minute).Err()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: "redis unhealthy"})
		}

		return c.JSON(http.StatusOK, dto.PingResponse{Message: "pong"})
	}
}
