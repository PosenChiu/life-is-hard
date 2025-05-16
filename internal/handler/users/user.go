// File: internal/handler/users/users.go
package users

import (
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"life-is-hard/internal/dto"
	"life-is-hard/internal/middleware"
	"life-is-hard/internal/model"
	"life-is-hard/internal/repository"
	"life-is-hard/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// CreateUserHandler 建立新使用者
// @Summary     Create a new user
// @Description 接收使用者表單資料並建立新帳號 (Email 會自動轉小寫)
// @Tags        users
// @Accept      application/x-www-form-urlencoded
// @Produce     json
// @Param       name     formData string true  "使用者姓名"
// @Param       email    formData string true  "使用者 Email (lowercase)"
// @Param       password formData string true  "使用者密碼"
// @Param       is_admin formData boolean true  "是否為管理員"
// @Success     201      {object} dto.UserResponse
// @Failure     400      {object} dto.ErrorResponse
// @Failure     500      {object} dto.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users [post]
func CreateUserHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreateUserRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "invalid form data"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: err.Error()})
		}

		hash, err := service.HashPassword(req.Password)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "failed to hash password"})
		}

		req.Email = strings.ToLower(req.Email)
		if _, err := mail.ParseAddress(req.Email); err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "invalid email format"})
		}

		user, err := repository.CreateUser(c.Request().Context(), pool, &model.User{
			Name:         req.Name,
			Email:        req.Email,
			PasswordHash: hash,
			IsAdmin:      req.IsAdmin,
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		}

		return c.JSON(http.StatusCreated, dto.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			IsAdmin:   user.IsAdmin,
		})
	}
}

// GetUserHandler 透過使用者 ID 取得使用者資訊
// @Summary     Get a user by ID
// @Description 透過 ID 查詢並回傳使用者詳細資料
// @Tags        users
// @Produce     json
// @Param       user_id   path      int  true  "使用者 ID"
// @Success     200  {object}  dto.UserResponse
// @Failure     400  {object}  dto.ErrorResponse  "參數錯誤"
// @Failure     404  {object}  dto.ErrorResponse  "使用者不存在"
// @Failure     500  {object}  dto.ErrorResponse  "伺服器錯誤"
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/{user_id} [get]
func GetUserHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("user_id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "invalid user ID"})
		}
		user, err := repository.GetUserByID(c.Request().Context(), pool, id)
		if err != nil {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Message: "user not found"})
		}
		return c.JSON(http.StatusOK, dto.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			IsAdmin:   user.IsAdmin,
		})
	}
}

// UpdateUserHandler 更新指定使用者資料
// @Summary     Update a user by ID
// @Description 根據使用者 ID 更新使用者姓名、Email 及管理員狀態
// @Tags        users
// @Accept      application/x-www-form-urlencoded
// @Produce     json
// @Param       user_id  path     int    true  "使用者 ID"
// @Param       name     formData string true  "使用者姓名"
// @Param       email    formData string true  "使用者 Email (lowercase)"
// @Param       is_admin formData boolean true "是否為管理員"
// @Success     204      "No Content"
// @Failure     400      {object} dto.ErrorResponse
// @Failure     404      {object} dto.ErrorResponse
// @Failure     500      {object} dto.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/{user_id} [put]
func UpdateUserHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "invalid user ID"})
		}

		var req dto.UpdateUserRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "invalid form data"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: err.Error()})
		}

		req.Email = strings.ToLower(req.Email)
		if _, err := mail.ParseAddress(req.Email); err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "invalid email format"})
		}

		if err := repository.UpdateUser(c.Request().Context(), pool, &model.User{
			ID:    id,
			Name:  req.Name,
			Email: req.Email,
		}); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		}

		return c.NoContent(http.StatusNoContent)
	}
}

// DeleteUserHandler 刪除指定 ID 的使用者
// @Summary     Delete a user by ID
// @Description 根據使用者 ID 刪除使用者帳號
// @Tags        users
// @Param       user_id   path      int  true  "使用者 ID"
// @Success     204  "No Content"
// @Failure     400  {object}  dto.ErrorResponse  "參數錯誤"
// @Failure     500  {object}  dto.ErrorResponse  "伺服器錯誤"
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/{user_id} [delete]
func DeleteUserHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("user_id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "invalid user ID"})
		}
		if err := repository.DeleteUser(c.Request().Context(), pool, id); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		}
		return c.NoContent(http.StatusNoContent)
	}
}

// GetMyUserHandler 取得當前使用者資訊
// @Summary     Get current user info
// @Description 透過 JWT Token 取得當前使用者詳細資訊
// @Tags        users
// @Produce     json
// @Success     200 {object} dto.UserResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me [get]
func GetMyUserHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get(middleware.ContextUserKey).(*service.CustomClaims)
		if !ok || claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "invalid or missing token"})
		}
		user, err := repository.GetUserByID(c.Request().Context(), pool, claims.UserID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		}
		return c.JSON(http.StatusOK, dto.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			IsAdmin:   user.IsAdmin,
		})
	}
}

// UpdateMyUserHandler 更新當前使用者資料
// @Summary     Update current user info
// @Description 使用 JWT 更新當前使用者姓名和 Email
// @Tags        users
// @Accept      application/x-www-form-urlencoded
// @Produce     json
// @Param       name  formData string true "使用者姓名"
// @Param       email formData string true "使用者 Email (lowercase)"
// @Success     204   "No Content"
// @Failure     400   {object} dto.ErrorResponse
// @Failure     401   {object} dto.ErrorResponse
// @Failure     500   {object} dto.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me [put]
func UpdateMyUserHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.UpdateUserRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "invalid form data"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: err.Error()})
		}

		claims, ok := c.Get(middleware.ContextUserKey).(*service.CustomClaims)
		if !ok || claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "invalid or missing token"})
		}

		req.Email = strings.ToLower(req.Email)
		if _, err := mail.ParseAddress(req.Email); err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "invalid email format"})
		}

		err := repository.UpdateUser(c.Request().Context(), pool, &model.User{
			ID:    claims.UserID,
			Name:  req.Name,
			Email: req.Email,
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		}

		return c.NoContent(http.StatusNoContent)
	}
}

// UpdateMyUserPasswordHandler 更新當前使用者密碼
// @Summary     Update own password
// @Description 驗證舊密碼並更新為新密碼
// @Tags        users
// @Accept      application/x-www-form-urlencoded
// @Produce     json
// @Param       old_password formData string true "當前密碼"
// @Param       new_password formData string true "新密碼"
// @Success     204      "No Content"
// @Failure     400      {object} dto.ErrorResponse
// @Failure     401      {object} dto.ErrorResponse
// @Failure     500      {object} dto.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me/password [patch]
func UpdateMyUserPasswordHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.UpdateMyPasswordRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "invalid form data"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: err.Error()})
		}

		claims, ok := c.Get(middleware.ContextUserKey).(*service.CustomClaims)
		if !ok || claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "invalid or missing token"})
		}

		user, err := repository.GetUserByID(c.Request().Context(), pool, claims.UserID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		}

		if err := service.AuthenticateUser(c.Request().Context(), *user, req.OldPassword); err != nil {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "invalid current password"})
		}

		hash, err := service.HashPassword(req.NewPassword)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: "failed to hash new password"})
		}

		if err := repository.UpdateUserPassword(c.Request().Context(), pool, claims.UserID, hash); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		}

		return c.NoContent(http.StatusNoContent)
	}
}

// DeleteMyUserHandler 刪除當前使用者帳號
// @Summary     Delete current user
// @Description 使用 JWT Token 刪除當前使用者帳號
// @Tags        users
// @Produce     json
// @Success     204
// @Failure     401 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me [delete]
func DeleteMyUserHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get(middleware.ContextUserKey).(*service.CustomClaims)
		if !ok || claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "invalid or missing token"})
		}
		if err := repository.DeleteUser(c.Request().Context(), pool, claims.UserID); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		}
		return c.NoContent(http.StatusNoContent)
	}
}
