// File: internal/handler/oauth/oauth_client.go

package oauth

import (
	"errors"
	"net/http"
	"strconv"

	"life-is-hard/internal/dto"
	"life-is-hard/internal/model"
	"life-is-hard/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// toResponse 將 model 轉成 API 回應
func toResponse(c *model.OAuthClient) dto.OAuthClientResponse {
	return dto.OAuthClientResponse{
		ID:           c.ID,
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		OwnerID:      c.OwnerID,
		GrantTypes:   c.GrantTypes,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}

// CreateOAuthClientHandler 新增 OAuth client
// @Summary     Create a new OAuth client
// @Description 僅支援 password 與 client_credentials
// @Tags        oauth
// @Accept      json
// @Produce     json
// @Param       request body dto.CreateOAuthClientRequest true "Create OAuth client request"
// @Success     201     {object} dto.OAuthClientResponse
// @Failure     400     {object} dto.HTTPError
// @Failure     500     {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /oauth/clients [post]
func CreateOAuthClientHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreateOAuthClientRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: err.Error()})
		}

		oc := &model.OAuthClient{
			ClientID:     req.ClientID,
			ClientSecret: req.ClientSecret,
			OwnerID:      req.OwnerID,
			GrantTypes:   req.GrantTypes,
		}
		if err := repository.CreateOAuthClient(c.Request().Context(), db, oc); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}
		return c.JSON(http.StatusCreated, toResponse(oc))
	}
}

// ListOAuthClientsHandler 列出所有 OAuth clients
// @Summary     List OAuth clients
// @Description 取得所有 OAuth clients
// @Tags        oauth
// @Produce     json
// @Success     200 {array} dto.OAuthClientResponse
// @Failure     500 {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /oauth/clients [get]
func ListOAuthClientsHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		clients, err := repository.ListOAuthClients(c.Request().Context(), db)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}
		var resp []dto.OAuthClientResponse
		for i := range clients {
			resp = append(resp, toResponse(&clients[i]))
		}
		return c.JSON(http.StatusOK, resp)
	}
}

// GetOAuthClientHandler 取得指定 ID 的 OAuth client
// @Summary     Get OAuth client
// @Description 根據 ID 取得 OAuth client
// @Tags        oauth
// @Produce     json
// @Param       id   path int true "OAuth client ID"
// @Success     200  {object} dto.OAuthClientResponse
// @Failure     400  {object} dto.HTTPError
// @Failure     404  {object} dto.HTTPError
// @Failure     500  {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /oauth/clients/{id} [get]
func GetOAuthClientHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid id"})
		}
		oc, err := repository.GetOAuthClientByID(c.Request().Context(), db, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return c.JSON(http.StatusNotFound, dto.HTTPError{Message: "OAuth client not found"})
			}
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}
		return c.JSON(http.StatusOK, toResponse(oc))
	}
}

// UpdateOAuthClientHandler 更新指定 ID 的 OAuth client
// @Summary     Update OAuth client
// @Tags        oauth
// @Accept      json
// @Produce     json
// @Param       id      path int true "OAuth client ID"
// @Param       request body dto.UpdateOAuthClientRequest true "Update OAuth client request"
// @Success     200     {object} dto.OAuthClientResponse
// @Failure     400     {object} dto.HTTPError
// @Failure     404     {object} dto.HTTPError
// @Failure     500     {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /oauth/clients/{id} [put]
func UpdateOAuthClientHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid id"})
		}
		var req dto.UpdateOAuthClientRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: err.Error()})
		}

		oc := &model.OAuthClient{
			ID:           id,
			ClientSecret: req.ClientSecret,
			OwnerID:      req.OwnerID,
			GrantTypes:   req.GrantTypes,
		}
		if err := repository.UpdateOAuthClient(c.Request().Context(), db, oc); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return c.JSON(http.StatusNotFound, dto.HTTPError{Message: "OAuth client not found"})
			}
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}
		return c.JSON(http.StatusOK, toResponse(oc))
	}
}

// DeleteOAuthClientHandler 刪除指定 ID 的 OAuth client
// @Summary     Delete OAuth client
// @Tags        oauth
// @Produce     json
// @Param       id   path int true "OAuth client ID"
// @Success     204
// @Failure     400  {object} dto.HTTPError
// @Failure     500  {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /oauth/clients/{id} [delete]
func DeleteOAuthClientHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid id"})
		}
		if err := repository.DeleteOAuthClient(c.Request().Context(), db, id); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}
		return c.NoContent(http.StatusNoContent)
	}
}
