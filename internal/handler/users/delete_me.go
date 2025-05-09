// File: internal/handler/users/delete_me.go
package users

import (
	"net/http"

	"life-is-hard/internal/dto"
	"life-is-hard/internal/repository"
	"life-is-hard/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// DeleteMeHandler 刪除當前使用者帳號
// @Summary     Delete current user
// @Description 使用 JWT Token 刪除當前使用者帳號
// @Tags        users
// @Produce     json
// @Success     204
// @Failure     401 {object} dto.HTTPError
// @Failure     500 {object} dto.HTTPError
// @Security    OAuth2Password[default]
// @Router      /users/me [delete]
func DeleteMeHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 從 context 中取出 JWT claims
		claimsRaw := c.Get("user")
		claims, ok := claimsRaw.(*service.CustomClaims)
		if !ok || claimsRaw == nil {
			return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid or missing token"})
		}

		// 刪除當前使用者
		if err := repository.DeleteUser(c.Request().Context(), pool, claims.ID); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}

		// 回傳 No Content
		return c.NoContent(http.StatusNoContent)
	}
}
