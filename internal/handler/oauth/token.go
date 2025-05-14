package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"life-is-hard/internal/dto"
	"life-is-hard/internal/model"
	"life-is-hard/internal/repository"
	"life-is-hard/internal/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// TokenRequest defines the expected form fields for the OAuth2 token endpoint request.
// swagger:model TokenRequest
type TokenRequest struct {
	// OAuth2 grant type: "password", "client_credentials", or "refresh_token"
	// required: true
	GrantType string `form:"grant_type" validate:"required" example:"password"`
	// Username (required if grant_type=password)
	Username string `form:"username" example:"alice"`
	// Password (required if grant_type=password)
	Password string `form:"password" example:"Secret123!"`
	// Refresh token (required if grant_type=refresh_token)
	RefreshToken string `form:"refresh_token" example:"I6gwLg8K..."`
	// Scope (optional) – space or comma separated
	Scope string `form:"scope" example:"read write"`

	// ClientID and ClientSecret are extracted from the Authorization header.
	ClientID     string `swaggerignore:"true"`
	ClientSecret string `swaggerignore:"true"`
}

// TokenResponse defines the JSON output format of the token endpoint.
// swagger:model TokenResponse
type TokenResponse struct {
	// JWT access token
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	// Token type (always "Bearer")
	TokenType string `json:"token_type" example:"Bearer"`
	// Expiration time in seconds
	ExpiresIn int `json:"expires_in" example:"86400"`
	// Refresh token to obtain new access tokens
	RefreshToken string `json:"refresh_token,omitempty" example:"I6gwLg8KGE2oIyJVwQc8fw..."`
	// Scope of the access token
	Scope string `json:"scope" example:"read write"`
}

// RefreshTokenData is the structure stored in Redis for each refresh token
type RefreshTokenData struct {
	UserID   int    `json:"user_id"`
	ClientID string `json:"client_id"`
	Scope    string `json:"scope"`
}

// TokenHandler handles the OAuth2 token endpoint (POST /api/oauth/token).
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
// @Param       scope          formData string false "Requested scope (optional)"
// @Success     200 {object} TokenResponse
// @Failure     400 {object} dto.HTTPError
// @Failure     401 {object} dto.HTTPError
// @Failure     500 {object} dto.HTTPError
// @Router      /oauth/token [post]
func TokenHandler(db *pgxpool.Pool, rdb *redis.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		var req TokenRequest

		// Bind and validate form input (excluding client credentials)
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: err.Error()})
		}

		// --- parse client credentials from Authorization header ---
		authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "missing authorization header"})
		}
		const prefix = "Basic "
		if !strings.HasPrefix(authHeader, prefix) {
			return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid authorization header"})
		}
		decoded, err := base64.StdEncoding.DecodeString(authHeader[len(prefix):])
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid authorization header"})
		}
		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) != 2 {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid authorization header"})
		}
		req.ClientID = parts[0]
		req.ClientSecret = parts[1]
		// ------------------------------------------------------------

		// Normalize and check grant_type
		grantType := strings.ToLower(req.GrantType)
		if grantType != "password" && grantType != "client_credentials" && grantType != "refresh_token" {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "unsupported grant_type"})
		}
		// Ensure required fields for each grant type
		if grantType == "password" {
			if req.Username == "" || req.Password == "" {
				return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid request"})
			}
		}
		if grantType == "refresh_token" {
			if req.RefreshToken == "" {
				return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid request"})
			}
		}

		// Authenticate the client using client_id and client_secret
		oc, err := repository.GetOAuthClientByClientID(ctx, db, req.ClientID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid client"})
			}
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}
		if oc.ClientSecret != req.ClientSecret {
			return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid client"})
		}
		// Check if client is allowed to use this grant type
		if grantType != "refresh_token" {
			allowed := false
			for _, g := range oc.GrantTypes {
				if g == grantType {
					allowed = true
					break
				}
			}
			if !allowed {
				return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "unauthorized grant_type"})
			}
		}
		// Prepare for token issuance
		var user *model.User
		scope := req.Scope // requested scope (could be empty or not)
		if grantType == "password" {
			// Verify user credentials
			u, err := repository.GetUserByName(ctx, db, req.Username)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid credentials"})
			}
			authUser, err := service.AuthenticateUser(ctx, *u, req.Password)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid credentials"})
			}
			user = authUser
		}
		if grantType == "refresh_token" {
			// Validate the provided refresh token via Redis
			key := "refresh_token:" + req.RefreshToken
			val, err := rdb.Get(ctx, key).Result()
			if err != nil {
				if err == redis.Nil {
					return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid refresh token"})
				}
				return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: fmt.Sprintf("failed to verify refresh token: %v", err)})
			}
			var rtData RefreshTokenData
			if jsonErr := json.Unmarshal([]byte(val), &rtData); jsonErr != nil {
				return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: "failed to verify refresh token"})
			}
			// Ensure the refresh token belongs to the same client
			if rtData.ClientID != req.ClientID {
				return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid refresh token"})
			}
			// If the token is tied to a user, fetch the latest user data
			if rtData.UserID != 0 {
				u, err := repository.GetUserByID(ctx, db, rtData.UserID)
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid refresh token"})
					}
					return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
				}
				user = u
			}
			// Use the scope from the original token
			scope = rtData.Scope
		}
		// Issue a new JWT access token
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: "failed to issue token: configuration error"})
		}
		now := time.Now()
		expiresAt := now.Add(24 * time.Hour)
		claims := jwt.MapClaims{
			"iss":       "life-is-hard",
			"exp":       expiresAt.Unix(),
			"client_id": oc.ClientID,
		}
		if user != nil {
			// Token on behalf of a user
			claims["sub"] = fmt.Sprint(user.ID)
			claims["id"] = user.ID
			claims["is_admin"] = user.IsAdmin
		} else {
			// Token for client itself (no user)
			claims["sub"] = oc.ClientID
		}
		if scope != "" {
			claims["scope"] = scope
		}
		tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := tokenObj.SignedString([]byte(secret))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: fmt.Sprintf("failed to issue token: %v", err)})
		}
		// Generate a refresh token if appropriate
		var newRefreshToken string
		if grantType == "password" {
			// Create a new random refresh token and store it
			if newRefreshToken, err = generateRandomToken(32); err != nil {
				return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: fmt.Sprintf("failed to issue token: %v", err)})
			}
			rtData := RefreshTokenData{
				UserID:   user.ID,
				ClientID: oc.ClientID,
				Scope:    scope,
			}
			dataBytes, _ := json.Marshal(rtData)
			if err := rdb.Set(ctx, "refresh_token:"+newRefreshToken, dataBytes, 24*30*time.Hour).Err(); err != nil {
				return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: fmt.Sprintf("failed to issue token: %v", err)})
			}
		}
		// Build response
		resp := TokenResponse{
			AccessToken: tokenStr,
			TokenType:   "Bearer",
			ExpiresIn:   int(time.Until(expiresAt).Seconds()),
			Scope:       scope,
		}
		if newRefreshToken != "" {
			resp.RefreshToken = newRefreshToken
		}
		return c.JSON(http.StatusOK, resp)
	}
}

// generateRandomToken creates a URL-safe random token string of n bytes
func generateRandomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// Encode to URL-safe base64 (no padding)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}
