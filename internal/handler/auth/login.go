// File: internal/handler/auth/login.go
package auth

import (
	"fmt"
	"net/http"
	"time"

	"life-is-hard/internal/dto"
	"life-is-hard/internal/repository"
	"life-is-hard/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// LoginHandler 使用 Username/Password 驗證並回傳 JWT
// @Summary     登入使用者
// @Description 使用 Username 與 Password 進行驗證，回傳存取令牌與到期時間
// @Tags        auth
// @Accept      application/x-www-form-urlencoded
// @Produce     json
// @Param       username formData string true "使用者名稱"
// @Param       password formData string true "使用者密碼"
// @Success     200      {object} dto.LoginResponse
// @Failure     400      {object} dto.ErrorResponse
// @Failure     401      {object} dto.ErrorResponse
// @Failure     500      {object} dto.ErrorResponse
// @Router      /auth/login [post]
func LoginHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.LoginRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: fmt.Sprintf("無效的表單資料: %v", err)})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: err.Error()})
		}

		user, err := repository.GetUserByName(c.Request().Context(), pool, req.Username)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "invalid credentials"})
		}
		if err := service.AuthenticateUser(c.Request().Context(), *user, req.Password); err != nil {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "invalid credentials"})
		}

		token, err := service.IssueAccessToken(*user, 24*time.Hour)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: fmt.Sprintf("failed to issue token: %v", err)})
		}

		return c.JSON(http.StatusOK, dto.LoginResponse{AccessToken: token})
	}
}
