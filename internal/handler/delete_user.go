// File: internal/handler/delete_user.go
package handler

import (
	"net/http"
	"strconv"

	"life-is-hard/internal/dto"
	"life-is-hard/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// DeleteUserHandler 刪除指定 ID 的使用者
// @Summary     Delete a user by ID
// @Description 根據使用者 ID 刪除使用者帳號
// @Tags        users
// @Param       id   path      int  true  "使用者 ID"
// @Success     204  "No Content"
// @Failure     400  {object}  dto.HTTPError  "參數錯誤"
// @Failure     500  {object}  dto.HTTPError  "伺服器錯誤"
// @Security    OAuth2Password[default]
// @Router      /users/{id} [delete]
func DeleteUserHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 解析 path 參數
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid user ID"})
		}

		// 執行刪除操作
		if err := repository.DeleteUser(c.Request().Context(), pool, id); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}

		// 回傳 204 No Content
		return c.NoContent(http.StatusNoContent)
	}
}
