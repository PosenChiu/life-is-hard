// File: internal/handler/users/oauth_client.go

package users

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"life-is-hard/internal/dto"
	"life-is-hard/internal/model"
	"life-is-hard/internal/repository"
	"life-is-hard/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// ----------
// DTOs
// ----------

// CreateUserOAuthClientRequest payload to create a client for the authenticated user.
// swagger:model CreateUserOAuthClientRequest
type CreateUserOAuthClientRequest struct {
	// required: true
	ClientID string `json:"client_id" validate:"required" example:"my-client"`
	// required: true
	ClientSecret string `json:"client_secret" validate:"required" example:"secret"`
	// required: true
	// 授權類型，逗號分隔 (password,client_credentials)
	GrantTypes []string `json:"grant_types" validate:"required" example:"password,client_credentials"`
}

// UpdateUserOAuthClientRequest payload to update a client for the authenticated user.
// swagger:model UpdateUserOAuthClientRequest
type UpdateUserOAuthClientRequest struct {
	// required: true
	ClientSecret string `json:"client_secret" validate:"required" example:"new-secret"`
	// required: true
	// 授權類型，逗號分隔 (password)
	GrantTypes []string `json:"grant_types" validate:"required" example:"password"`
}

// OAuthClientResponse API response for OAuth client.
// swagger:model OAuthClientResponse
type OAuthClientResponse struct {
	ID           int    `json:"id" example:"1"`
	ClientID     string `json:"client_id" example:"my-client"`
	ClientSecret string `json:"client_secret" example:"secret"`
	OwnerID      int    `json:"owner_id" example:"42"`
	// 授權類型，逗號分隔 (password,client_credentials)
	GrantTypes []string  `json:"grant_types" example:"password,client_credentials"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// convertToResponse maps model to DTO.
func convertToResponse(c *model.OAuthClient) OAuthClientResponse {
	return OAuthClientResponse{
		ID:           c.ID,
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		OwnerID:      c.OwnerID,
		GrantTypes:   c.GrantTypes,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}

// ----------
// Handlers
// ----------

// CreateUserOAuthClientHandler handles POST /users/me/oauth/clients
// @Summary     Create OAuth client for authenticated user
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       request body CreateUserOAuthClientRequest true "Create OAuth client"
// @Success     201 {object} OAuthClientResponse
// @Failure     400 {object} dto.HTTPError
// @Failure     401 {object} dto.HTTPError
// @Failure     500 {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me/oauth/clients [post]
func CreateUserOAuthClientHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		// get user ID from JWT claims
		claimsRaw := c.Get("user")
		claims, ok := claimsRaw.(*service.CustomClaims)
		if !ok || claimsRaw == nil {
			return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid or missing token"})
		}
		userID := claims.UserID

		// bind and validate
		var req CreateUserOAuthClientRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: err.Error()})
		}

		// create and persist model
		oc := &model.OAuthClient{
			ClientID:     req.ClientID,
			ClientSecret: req.ClientSecret,
			OwnerID:      userID,
			GrantTypes:   req.GrantTypes,
		}
		if err := repository.CreateOAuthClient(c.Request().Context(), db, oc); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}
		return c.JSON(http.StatusCreated, convertToResponse(oc))
	}
}

// ListUserOAuthClientsHandler handles GET /users/me/oauth/clients
// @Summary     List OAuth clients for authenticated user
// @Tags        users
// @Accept      json
// @Produce     json
// @Success     200 {array} OAuthClientResponse
// @Failure     401 {object} dto.HTTPError
// @Failure     500 {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me/oauth/clients [get]
func ListUserOAuthClientsHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		claimsRaw := c.Get("user")
		claims, ok := claimsRaw.(*service.CustomClaims)
		if !ok || claimsRaw == nil {
			return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid or missing token"})
		}
		userID := claims.UserID

		all, err := repository.ListOAuthClients(c.Request().Context(), db)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}
		var resp []OAuthClientResponse
		for _, oc := range all {
			if oc.OwnerID == userID {
				resp = append(resp, convertToResponse(&oc))
			}
		}
		return c.JSON(http.StatusOK, resp)
	}
}

// GetUserOAuthClientHandler handles GET /users/me/oauth/clients/{client_id}
// @Summary     Get OAuth client for authenticated user
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       client_id path int true "Client ID"
// @Success     200 {object} OAuthClientResponse
// @Failure     400 {object} dto.HTTPError
// @Failure     401 {object} dto.HTTPError
// @Failure     404 {object} dto.HTTPError
// @Failure     500 {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me/oauth/clients/{client_id} [get]
func GetUserOAuthClientHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		claimsRaw := c.Get("user")
		claims, ok := claimsRaw.(*service.CustomClaims)
		if !ok || claimsRaw == nil {
			return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid or missing token"})
		}
		userID := claims.UserID

		clientID, err := strconv.Atoi(c.Param("client_id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid client_id"})
		}
		oc, err := repository.GetOAuthClientByID(c.Request().Context(), db, clientID)
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, dto.HTTPError{Message: "client not found"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}
		if oc.OwnerID != userID {
			return c.JSON(http.StatusNotFound, dto.HTTPError{Message: "client not found"})
		}
		return c.JSON(http.StatusOK, convertToResponse(oc))
	}
}

// UpdateUserOAuthClientHandler handles PUT /users/me/oauth/clients/{client_id}
// @Summary     Update OAuth client for authenticated user
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       client_id path int true "Client ID"
// @Param       request   body UpdateUserOAuthClientRequest true "Update OAuth client"
// @Success     200 {object} OAuthClientResponse
// @Failure     400 {object} dto.HTTPError
// @Failure     401 {object} dto.HTTPError
// @Failure     404 {object} dto.HTTPError
// @Failure     500 {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me/oauth/clients/{client_id} [put]
func UpdateUserOAuthClientHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		claimsRaw := c.Get("user")
		claims, ok := claimsRaw.(*service.CustomClaims)
		if !ok || claimsRaw == nil {
			return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid or missing token"})
		}
		userID := claims.UserID

		clientID, err := strconv.Atoi(c.Param("client_id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid client_id"})
		}

		var req UpdateUserOAuthClientRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: err.Error()})
		}

		oc, err := repository.GetOAuthClientByID(c.Request().Context(), db, clientID)
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, dto.HTTPError{Message: "client not found"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}
		if oc.OwnerID != userID {
			return c.JSON(http.StatusNotFound, dto.HTTPError{Message: "client not found"})
		}

		oc.ClientSecret = req.ClientSecret
		oc.GrantTypes = req.GrantTypes
		oc.UpdatedAt = time.Now().UTC()

		if err := repository.UpdateOAuthClient(c.Request().Context(), db, oc); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}
		return c.JSON(http.StatusOK, convertToResponse(oc))
	}
}

// DeleteUserOAuthClientHandler handles DELETE /users/me/oauth/clients/{client_id}
// @Summary     Delete OAuth client for authenticated user
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       client_id path int true "Client ID"
// @Success     204
// @Failure     400 {object} dto.HTTPError
// @Failure     401 {object} dto.HTTPError
// @Failure     404 {object} dto.HTTPError
// @Failure     500 {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me/oauth/clients/{client_id} [delete]
func DeleteUserOAuthClientHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		claimsRaw := c.Get("user")
		claims, ok := claimsRaw.(*service.CustomClaims)
		if !ok || claimsRaw == nil {
			return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid or missing token"})
		}
		userID := claims.UserID

		clientID, err := strconv.Atoi(c.Param("client_id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid client_id"})
		}

		oc, err := repository.GetOAuthClientByID(c.Request().Context(), db, clientID)
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, dto.HTTPError{Message: "client not found"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}
		if oc.OwnerID != userID {
			return c.JSON(http.StatusNotFound, dto.HTTPError{Message: "client not found"})
		}

		if err := repository.DeleteOAuthClient(c.Request().Context(), db, clientID); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}
		return c.NoContent(http.StatusNoContent)
	}
}
