package users

import (
	"net/http"
	"time"

	"life-is-hard/internal/api"
	"life-is-hard/internal/database"
	"life-is-hard/internal/middleware"
	"life-is-hard/internal/model"
	"life-is-hard/internal/service"
	"life-is-hard/internal/store"

	"github.com/labstack/echo/v4"
)

// @Summary     Create OAuth client for authenticated user
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       request body api.CreateOAuthClientRequest true "Create OAuth client"
// @Success     201 {object} api.OAuthClientResponse
// @Failure     400 {object} api.ErrorResponse
// @Failure     401 {object} api.ErrorResponse
// @Failure     500 {object} api.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me/oauth-clients [post]
func CreateMyOAuthClientHandler(db database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get(middleware.ContextUserKey).(*service.CustomClaims)
		if !ok || claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, api.ErrorResponse{Message: "invalid or missing token"})
		}

		var req api.CreateOAuthClientRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: err.Error()})
		}

		client := &model.OAuthClient{
			ClientID:     req.ClientID,
			ClientSecret: req.ClientSecret,
			UserID:       claims.UserID,
			GrantTypes:   req.GrantTypes,
		}
		if err := store.CreateOAuthClient(c.Request().Context(), db, client); err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		}
		return c.JSON(http.StatusCreated, api.OAuthClientResponse{
			ClientID:     client.ClientID,
			ClientSecret: client.ClientSecret,
			UserID:       client.UserID,
			GrantTypes:   client.GrantTypes,
			CreatedAt:    client.CreatedAt,
			UpdatedAt:    client.UpdatedAt,
		})
	}
}

// @Summary     List OAuth clients for authenticated user
// @Tags        users
// @Accept      json
// @Produce     json
// @Success     200 {array} api.OAuthClientResponse
// @Failure     401 {object} api.ErrorResponse
// @Failure     500 {object} api.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me/oauth-clients [get]
func ListMyOAuthClientsHandler(db database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get(middleware.ContextUserKey).(*service.CustomClaims)
		if !ok || claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, api.ErrorResponse{Message: "invalid or missing token"})
		}

		clients, err := store.ListOAuthClients(c.Request().Context(), db, claims.UserID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		}

		resp := make([]api.OAuthClientResponse, len(clients))
		for i, client := range clients {
			resp[i] = api.OAuthClientResponse{
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

// @Summary     Get OAuth client for authenticated user
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       client_id path int true "Client ID"
// @Success     200 {object} api.OAuthClientResponse
// @Failure     400 {object} api.ErrorResponse
// @Failure     401 {object} api.ErrorResponse
// @Failure     404 {object} api.ErrorResponse
// @Failure     500 {object} api.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me/oauth-clients/{client_id} [get]
func GetMyOAuthClientHandler(db database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get(middleware.ContextUserKey).(*service.CustomClaims)
		if !ok || claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, api.ErrorResponse{Message: "invalid or missing token"})
		}

		client, err := store.GetOAuthClientByClientID(c.Request().Context(), db, c.Param("client_id"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		}
		if client.UserID != claims.UserID {
			return c.JSON(http.StatusNotFound, api.ErrorResponse{Message: "client not found"})
		}

		return c.JSON(http.StatusOK, api.OAuthClientResponse{
			ClientID:     client.ClientID,
			ClientSecret: client.ClientSecret,
			UserID:       client.UserID,
			GrantTypes:   client.GrantTypes,
			CreatedAt:    client.CreatedAt,
			UpdatedAt:    client.UpdatedAt,
		})
	}
}

// @Summary     Update OAuth client for authenticated user
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       client_id path int true "Client ID"
// @Param       request   body api.UpdateOAuthClientRequest true "Update OAuth client"
// @Success     200 {object} api.OAuthClientResponse
// @Failure     400 {object} api.ErrorResponse
// @Failure     401 {object} api.ErrorResponse
// @Failure     404 {object} api.ErrorResponse
// @Failure     500 {object} api.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me/oauth-clients/{client_id} [put]
func UpdateMyOAuthClientHandler(db database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get(middleware.ContextUserKey).(*service.CustomClaims)
		if !ok || claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, api.ErrorResponse{Message: "invalid or missing token"})
		}

		var req api.UpdateOAuthClientRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: err.Error()})
		}

		client, err := store.GetOAuthClientByClientID(c.Request().Context(), db, c.Param("client_id"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		}
		if client.UserID != claims.UserID {
			return c.JSON(http.StatusNotFound, api.ErrorResponse{Message: "client not found"})
		}

		client.ClientSecret = req.ClientSecret
		client.GrantTypes = req.GrantTypes
		client.UpdatedAt = time.Now().UTC()

		if err := store.UpdateOAuthClient(c.Request().Context(), db, client); err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		}

		return c.JSON(http.StatusOK, api.OAuthClientResponse{
			ClientID:     client.ClientID,
			ClientSecret: client.ClientSecret,
			UserID:       client.UserID,
			GrantTypes:   client.GrantTypes,
			CreatedAt:    client.CreatedAt,
			UpdatedAt:    client.UpdatedAt,
		})
	}
}

// @Summary     Delete OAuth client for authenticated user
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       client_id path int true "Client ID"
// @Success     204
// @Failure     400 {object} api.ErrorResponse
// @Failure     401 {object} api.ErrorResponse
// @Failure     404 {object} api.ErrorResponse
// @Failure     500 {object} api.ErrorResponse
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me/oauth-clients/{client_id} [delete]
func DeleteMyOAuthClientHandler(db database.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get(middleware.ContextUserKey).(*service.CustomClaims)
		if !ok || claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, api.ErrorResponse{Message: "invalid or missing token"})
		}

		client, err := store.GetOAuthClientByClientID(c.Request().Context(), db, c.Param("client_id"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		}
		if client.UserID != claims.UserID {
			return c.JSON(http.StatusNotFound, api.ErrorResponse{Message: "client not found"})
		}

		if err := store.DeleteOAuthClient(c.Request().Context(), db, c.Param("client_id")); err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		}
		return c.NoContent(http.StatusNoContent)
	}
}
