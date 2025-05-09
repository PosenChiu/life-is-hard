// File: internal/handler/users/get_me.go
package users

import (
	"net/http"

	"life-is-hard/internal/dto"
	"life-is-hard/internal/repository"
	"life-is-hard/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// GetMeHandler 取得當前使用者資訊
// @Summary     Get current user info
// @Description 透過 JWT Token 取得當前使用者詳細資訊
// @Tags        users
// @Produce     json
// @Success     200 {object} UserResponse
// @Failure     401 {object} dto.HTTPError
// @Failure     500 {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Router      /users/me [get]
func GetMeHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 從 context 取得 JWT claims
		claimsRaw := c.Get("user")
		claims, ok := claimsRaw.(*service.CustomClaims)
		if !ok || claimsRaw == nil {
			return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid or missing token"})
		}

		// 根據 claims.ID 查詢使用者
		user, err := repository.GetUserByID(c.Request().Context(), pool, claims.ID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}

		// 組裝回應
		resp := UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			IsAdmin:   user.IsAdmin,
		}
		return c.JSON(http.StatusOK, resp)
	}
}
