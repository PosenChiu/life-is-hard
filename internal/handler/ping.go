// File: internal/handler/ping.go
package handler

import (
	"net/http"

	"life-is-hard/internal/dto"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// PingResponse 健康檢查回應模型
// swagger:model PingResponse
type PingResponse struct {
	// 回應訊息
	Message string `json:"message" example:"pong"`
}

// PingHandler 健康檢查（需通過認證）
// @Summary     Health Check
// @Description 回傳 pong，並檢查資料庫連線是否正常
// @Tags        health
// @Accept      json
// @Produce     json
// @Success     200 {object} PingResponse
// @Failure     500 {object} dto.HTTPError
// @Security    OAuth2Password[default]
// @Router      /ping [get]
func PingHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := db.Ping(c.Request().Context()); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: "database unhealthy"})
		}
		return c.JSON(http.StatusOK, PingResponse{Message: "pong"})
	}
}
