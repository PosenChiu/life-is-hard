package oauth

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"life-is-hard/internal/api"
	"life-is-hard/internal/cache"
	"life-is-hard/internal/database"
	"life-is-hard/internal/model"
	"life-is-hard/internal/service"
	"life-is-hard/internal/store"

	"github.com/labstack/echo/v4"
)

// 方便測試時取代底層依賴
var (
	getOAuthClientByClientID = store.GetOAuthClientByClientID
	getUserByName            = store.GetUserByName
	getUserByID              = store.GetUserByID
	authenticateUser         = service.AuthenticateUser
	issueAccessToken         = service.IssueAccessToken
	issueRefreshToken        = service.IssueRefreshToken
	issueClientAccessToken   = service.IssueClientAccessToken
	validateRefreshToken     = service.ValidateRefreshToken
)

// @Summary     OAuth2 obtain access token
// @Description Issue a JWT access token (and refresh token if applicable) using OAuth2 grant_type
// @Tags        oauth
// @Accept      application/x-www-form-urlencoded
// @Produce     json
// @Param       Authorization header string true  "Basic base64(client_id:client_secret)"
// @Param       grant_type     formData string true  "Grant type: password, client_credentials, or refresh_token"
// @Param       username       formData string false "Username (required for password grant)"
// @Param       password       formData string false "Password (required for password grant)"
// @Param       refresh_token  formData string false "Refresh token (required for refresh_token grant)"
// @Success     200 {object} api.TokenResponse
// @Failure     400 {object} api.ErrorResponse
// @Failure     401 {object} api.ErrorResponse
// @Failure     500 {object} api.ErrorResponse
// @Router      /oauth/token [post]
func TokenHandler(db database.DB, cache cache.Cache) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		var req api.TokenRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "invalid request payload"})
		}

		// 解析 Basic 認證
		auth := c.Request().Header.Get("Authorization")
		const prefix = "Basic "
		if !strings.HasPrefix(auth, prefix) {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "invalid authorization header"})
		}
		decoded, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
		if err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "invalid authorization header"})
		}
		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) != 2 {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "invalid authorization header"})
		}
		req.ClientID = parts[0]
		req.ClientSecret = parts[1]

		// 驗證 client
		oc, err := getOAuthClientByClientID(ctx, db, req.ClientID)
		if err != nil || oc.ClientSecret != req.ClientSecret {
			return c.JSON(http.StatusUnauthorized, api.ErrorResponse{Message: "invalid client credentials"})
		}

		// 檢查 grant_type
		allowed := false
		for _, gt := range oc.GrantTypes {
			if gt == req.GrantType {
				allowed = true
				break
			}
		}
		if !allowed {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "unauthorized grant_type"})
		}

		var tokenStr, newRefreshToken string

		switch req.GrantType {
		case "password":
			user, err := getUserByName(ctx, db, req.Username)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, api.ErrorResponse{Message: "invalid credentials"})
			}
			if err := authenticateUser(ctx, *user, req.Password); err != nil {
				return c.JSON(http.StatusUnauthorized, api.ErrorResponse{Message: "invalid credentials"})
			}

			// 發行 access token
			tokenStr, err = issueAccessToken(*user, 24*time.Hour)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: "failed to issue token"})
			}

			// 發行 refresh token
			newRefreshToken, err = issueRefreshToken(ctx, cache, user.ID, oc.ClientID, user.IsAdmin, 30*24*time.Hour)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: "failed to issue refresh token"})
			}

		case "client_credentials":
			// 為 client 自身（由 owner）發行 access token
			owner, err := getUserByID(ctx, db, oc.UserID)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: "failed to retrieve client owner"})
			}

			tokenStr, err = issueClientAccessToken(*owner, *oc, 24*time.Hour)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: "failed to issue token"})
			}

		case "refresh_token":
			// 驗證並讀取 refresh token
			data, err := validateRefreshToken(ctx, cache, req.RefreshToken)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, api.ErrorResponse{Message: "invalid refresh token"})
			}
			// 重新發行 access token
			tokenStr, err = issueAccessToken(model.User{ID: data.UserID, IsAdmin: false}, 24*time.Hour)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: "failed to issue token"})
			}
			// reuse same refresh token
			newRefreshToken = req.RefreshToken

		default:
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "unsupported grant_type"})
		}

		resp := api.TokenResponse{
			AccessToken:  tokenStr,
			TokenType:    "Bearer",
			ExpiresIn:    86400,
			RefreshToken: newRefreshToken,
		}
		return c.JSON(http.StatusOK, resp)
	}
}
