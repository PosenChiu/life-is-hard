// File: internal/handler/users/oauth_client.go

package users

import (
	"net/http"
	"time"

	"life-is-hard/internal/dto"
	"life-is-hard/internal/middleware"
	"life-is-hard/internal/model"
	"life-is-hard/internal/repository"
	"life-is-hard/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// CreateMyOAuthClientHandler handles POST /users/me/oauth/clients
// @Summary     Create OAuth client for authenticated user
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       request body dto.CreateOAuthClientRequest true "Create OAuth client"
// @Success     201 {object} dto.OAuthClientResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me/oauth/clients [post]
func CreateMyOAuthClientHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get(middleware.ContextUserKey).(*service.CustomClaims)
		if !ok || claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "invalid or missing token"})
		}

		var req dto.CreateOAuthClientRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: err.Error()})
		}

		client := &model.OAuthClient{
			ClientID:     req.ClientID,
			ClientSecret: req.ClientSecret,
			UserID:       claims.UserID,
			GrantTypes:   req.GrantTypes,
		}
		if err := repository.CreateOAuthClient(c.Request().Context(), db, client); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		}
		return c.JSON(http.StatusCreated, dto.OAuthClientResponse{
			ClientID:     client.ClientID,
			ClientSecret: client.ClientSecret,
			UserID:       client.UserID,
			GrantTypes:   client.GrantTypes,
			CreatedAt:    client.CreatedAt,
			UpdatedAt:    client.UpdatedAt,
		})
	}
}

// ListMyOAuthClientsHandler handles GET /users/me/oauth/clients
// @Summary     List OAuth clients for authenticated user
// @Tags        users
// @Accept      json
// @Produce     json
// @Success     200 {array} dto.OAuthClientResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me/oauth/clients [get]
func ListMyOAuthClientsHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get(middleware.ContextUserKey).(*service.CustomClaims)
		if !ok || claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "invalid or missing token"})
		}

		clients, err := repository.ListOAuthClients(c.Request().Context(), db, claims.UserID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		}

		resp := make([]dto.OAuthClientResponse, len(clients))
		for i, client := range clients {
			resp[i] = dto.OAuthClientResponse{
				ClientID:     client.ClientID,
				ClientSecret: client.ClientSecret,
				UserID:       client.UserID,
				GrantTypes:   client.GrantTypes,
				CreatedAt:    client.CreatedAt,
				UpdatedAt:    client.UpdatedAt,
			}
		}
		return c.JSON(http.StatusOK, resp)
	}
}

// GetMyOAuthClientHandler handles GET /users/me/oauth/clients/{client_id}
// @Summary     Get OAuth client for authenticated user
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       client_id path int true "Client ID"
// @Success     200 {object} dto.OAuthClientResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me/oauth/clients/{client_id} [get]
func GetMyOAuthClientHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get(middleware.ContextUserKey).(*service.CustomClaims)
		if !ok || claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "invalid or missing token"})
		}

		client, err := repository.GetOAuthClientByClientID(c.Request().Context(), db, c.Param("client_id"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		}
		if client.UserID != claims.UserID {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Message: "client not found"})
		}

		return c.JSON(http.StatusOK, dto.OAuthClientResponse{
			ClientID:     client.ClientID,
			ClientSecret: client.ClientSecret,
			UserID:       client.UserID,
			GrantTypes:   client.GrantTypes,
			CreatedAt:    client.CreatedAt,
			UpdatedAt:    client.UpdatedAt,
		})
	}
}

// UpdateMyOAuthClientHandler handles PUT /users/me/oauth/clients/{client_id}
// @Summary     Update OAuth client for authenticated user
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       client_id path int true "Client ID"
// @Param       request   body dto.UpdateOAuthClientRequest true "Update OAuth client"
// @Success     200 {object} dto.OAuthClientResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me/oauth/clients/{client_id} [put]
func UpdateMyOAuthClientHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get(middleware.ContextUserKey).(*service.CustomClaims)
		if !ok || claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "invalid or missing token"})
		}

		var req dto.UpdateOAuthClientRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: err.Error()})
		}

		client, err := repository.GetOAuthClientByClientID(c.Request().Context(), db, c.Param("client_id"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		}
		if client.UserID != claims.UserID {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Message: "client not found"})
		}

		client.ClientSecret = req.ClientSecret
		client.GrantTypes = req.GrantTypes
		client.UpdatedAt = time.Now().UTC()

		if err := repository.UpdateOAuthClient(c.Request().Context(), db, client); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		}

		return c.JSON(http.StatusOK, dto.OAuthClientResponse{
			ClientID:     client.ClientID,
			ClientSecret: client.ClientSecret,
			UserID:       client.UserID,
			GrantTypes:   client.GrantTypes,
			CreatedAt:    client.CreatedAt,
			UpdatedAt:    client.UpdatedAt,
		})
	}
}

// DeleteMyOAuthClientHandler handles DELETE /users/me/oauth/clients/{client_id}
// @Summary     Delete OAuth client for authenticated user
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       client_id path int true "Client ID"
// @Success     204
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me/oauth/clients/{client_id} [delete]
func DeleteMyOAuthClientHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get(middleware.ContextUserKey).(*service.CustomClaims)
		if !ok || claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "invalid or missing token"})
		}

		client, err := repository.GetOAuthClientByClientID(c.Request().Context(), db, c.Param("client_id"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		}
		if client.UserID != claims.UserID {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Message: "client not found"})
		}

		if err := repository.DeleteOAuthClient(c.Request().Context(), db, c.Param("client_id")); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		}
		return c.NoContent(http.StatusNoContent)
	}
}
