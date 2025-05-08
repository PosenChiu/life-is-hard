// File: internal/handler/get_user.go
package handler

import (
	"net/http"
	"strconv"

	"life-is-hard/internal/dto"
	"life-is-hard/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// GetUserHandler 透過使用者 ID 取得使用者資訊
// @Summary     Get a user by ID
// @Description 透過 ID 查詢並回傳使用者詳細資料
// @Tags        users
// @Produce     json
// @Param       id   path      int  true  "使用者 ID"
// @Success     200  {object}  UserResponse
// @Failure     400  {object}  dto.HTTPError  "參數錯誤"
// @Failure     404  {object}  dto.HTTPError  "使用者不存在"
// @Failure     500  {object}  dto.HTTPError  "伺服器錯誤"
// @Security    OAuth2Password[default]
// @Router      /users/{id} [get]
func GetUserHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 解析 path 參數
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid user ID"})
		}

		// 從資料庫讀取使用者資料
		user, err := repository.GetUserByID(c.Request().Context(), pool, id)
		if err != nil {
			return c.JSON(http.StatusNotFound, dto.HTTPError{Message: "user not found"})
		}

		// 組裝回傳用結構
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
