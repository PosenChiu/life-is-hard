package users

import (
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"life-is-hard/internal/api"
	"life-is-hard/internal/database"
	"life-is-hard/internal/middleware"
	"life-is-hard/internal/model"
	"life-is-hard/internal/service"
	"life-is-hard/internal/store"

	"github.com/labstack/echo/v4"
)

var (
	hashPassword       = service.HashPassword
	authenticateUser   = service.AuthenticateUser
	createUser         = store.CreateUser
	getUserByID        = store.GetUserByID
	updateUser         = store.UpdateUser
	updateUserPassword = store.UpdateUserPassword
	deleteUser         = store.DeleteUser
)

// @Summary     Create a new user
// @Description 接收使用者表單資料並建立新帳號 (Email 會自動轉小寫)
// @Tags        users
// @Accept      application/x-www-form-urlencoded
// @Produce     json
// @Param       name     formData string true  "使用者姓名"
// @Param       email    formData string true  "使用者 Email (lowercase)"
// @Param       password formData string true  "使用者密碼"
// @Param       is_admin formData boolean true  "是否為管理員"
// @Success     201      {object} api.UserResponse
// @Failure     400      {object} api.ErrorResponse
// @Failure     500      {object} api.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users [post]
func CreateUserHandler(db database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req api.CreateUserRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "invalid form data"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: err.Error()})
		}

		hash, err := hashPassword(req.Password)
		if err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "failed to hash password"})
		}

		req.Email = strings.ToLower(req.Email)
		if _, err := mail.ParseAddress(req.Email); err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "invalid email format"})
		}

		user, err := createUser(c.Request().Context(), db, &model.User{
			Name:         req.Name,
			Email:        req.Email,
			PasswordHash: hash,
			IsAdmin:      req.IsAdmin,
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		}

		return c.JSON(http.StatusCreated, api.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			IsAdmin:   user.IsAdmin,
		})
	}
}

// @Summary     Get a user by ID
// @Description 透過 ID 查詢並回傳使用者詳細資料
// @Tags        users
// @Produce     json
// @Param       user_id   path      int  true  "使用者 ID"
// @Success     200  {object}  api.UserResponse
// @Failure     400  {object}  api.ErrorResponse  "參數錯誤"
// @Failure     404  {object}  api.ErrorResponse  "使用者不存在"
// @Failure     500  {object}  api.ErrorResponse  "伺服器錯誤"
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/{user_id} [get]
func GetUserHandler(db database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("user_id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "invalid user ID"})
		}
		user, err := getUserByID(c.Request().Context(), db, id)
		if err != nil {
			return c.JSON(http.StatusNotFound, api.ErrorResponse{Message: "user not found"})
		}
		return c.JSON(http.StatusOK, api.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			IsAdmin:   user.IsAdmin,
		})
	}
}

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
// @Failure     400      {object} api.ErrorResponse
// @Failure     404      {object} api.ErrorResponse
// @Failure     500      {object} api.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/{user_id} [put]
func UpdateUserHandler(db database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "invalid user ID"})
		}

		var req api.UpdateUserRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "invalid form data"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: err.Error()})
		}

		req.Email = strings.ToLower(req.Email)
		if _, err := mail.ParseAddress(req.Email); err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "invalid email format"})
		}

		if err := updateUser(c.Request().Context(), db, &model.User{
			ID:    id,
			Name:  req.Name,
			Email: req.Email,
		}); err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		}

		return c.NoContent(http.StatusNoContent)
	}
}

// @Summary     Delete a user by ID
// @Description 根據使用者 ID 刪除使用者帳號
// @Tags        users
// @Param       user_id   path      int  true  "使用者 ID"
// @Success     204  "No Content"
// @Failure     400  {object}  api.ErrorResponse  "參數錯誤"
// @Failure     500  {object}  api.ErrorResponse  "伺服器錯誤"
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/{user_id} [delete]
func DeleteUserHandler(db database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("user_id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "invalid user ID"})
		}
		if err := deleteUser(c.Request().Context(), db, id); err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		}
		return c.NoContent(http.StatusNoContent)
	}
}

// @Summary     Get current user info
// @Description 透過 JWT Token 取得當前使用者詳細資訊
// @Tags        users
// @Produce     json
// @Success     200 {object} api.UserResponse
// @Failure     401 {object} api.ErrorResponse
// @Failure     500 {object} api.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me [get]
func GetMyUserHandler(db database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get(middleware.ContextUserKey).(*service.CustomClaims)
		if !ok || claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, api.ErrorResponse{Message: "invalid or missing token"})
		}
		user, err := getUserByID(c.Request().Context(), db, claims.UserID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		}
		return c.JSON(http.StatusOK, api.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			IsAdmin:   user.IsAdmin,
		})
	}
}

// @Summary     Update current user info
// @Description 使用 JWT 更新當前使用者姓名和 Email
// @Tags        users
// @Accept      application/x-www-form-urlencoded
// @Produce     json
// @Param       name  formData string true "使用者姓名"
// @Param       email formData string true "使用者 Email (lowercase)"
// @Success     204   "No Content"
// @Failure     400   {object} api.ErrorResponse
// @Failure     401   {object} api.ErrorResponse
// @Failure     500   {object} api.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me [put]
func UpdateMyUserHandler(db database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req api.UpdateUserRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "invalid form data"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: err.Error()})
		}

		claims, ok := c.Get(middleware.ContextUserKey).(*service.CustomClaims)
		if !ok || claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, api.ErrorResponse{Message: "invalid or missing token"})
		}

		req.Email = strings.ToLower(req.Email)
		if _, err := mail.ParseAddress(req.Email); err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "invalid email format"})
		}

		err := updateUser(c.Request().Context(), db, &model.User{
			ID:    claims.UserID,
			Name:  req.Name,
			Email: req.Email,
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		}

		return c.NoContent(http.StatusNoContent)
	}
}

// @Summary     Update own password
// @Description 驗證舊密碼並更新為新密碼
// @Tags        users
// @Accept      application/x-www-form-urlencoded
// @Produce     json
// @Param       old_password formData string true "當前密碼"
// @Param       new_password formData string true "新密碼"
// @Success     204      "No Content"
// @Failure     400      {object} api.ErrorResponse
// @Failure     401      {object} api.ErrorResponse
// @Failure     500      {object} api.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me/password [patch]
func UpdateMyUserPasswordHandler(db database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req api.UpdateMyPasswordRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "invalid form data"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: err.Error()})
		}

		claims, ok := c.Get(middleware.ContextUserKey).(*service.CustomClaims)
		if !ok || claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, api.ErrorResponse{Message: "invalid or missing token"})
		}

		user, err := getUserByID(c.Request().Context(), db, claims.UserID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		}

		if err := authenticateUser(c.Request().Context(), *user, req.OldPassword); err != nil {
			return c.JSON(http.StatusUnauthorized, api.ErrorResponse{Message: "invalid current password"})
		}

		hash, err := hashPassword(req.NewPassword)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: "failed to hash new password"})
		}

		if err := updateUserPassword(c.Request().Context(), db, claims.UserID, hash); err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		}

		return c.NoContent(http.StatusNoContent)
	}
}

// @Summary     Delete current user
// @Description 使用 JWT Token 刪除當前使用者帳號
// @Tags        users
// @Produce     json
// @Success     204
// @Failure     401 {object} api.ErrorResponse
// @Failure     500 {object} api.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me [delete]
func DeleteMyUserHandler(db database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get(middleware.ContextUserKey).(*service.CustomClaims)
		if !ok || claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, api.ErrorResponse{Message: "invalid or missing token"})
		}
		if err := deleteUser(c.Request().Context(), db, claims.UserID); err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		}
		return c.NoContent(http.StatusNoContent)
	}
}
