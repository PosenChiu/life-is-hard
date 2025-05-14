// File: internal/handler/oauth/oauth_client.go
package oauth

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"life-is-hard/internal/dto"
	"life-is-hard/internal/model"
	"life-is-hard/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// CreateOAuthClientRequest 建立 OAuth client 的請求
// swagger:model CreateOAuthClientRequest
type CreateOAuthClientRequest struct {
	// required: true
	ClientID string `json:"client_id" validate:"required" example:"my-client"`
	// required: true
	ClientSecret string `json:"client_secret" validate:"required" example:"secret"`
	// optional: 擁有者 user_id
	OwnerID *int `json:"owner_id" example:"1"`
	// required: 一般只需 ["password","client_credentials"]
	GrantTypes []string `json:"grant_types" validate:"required" example:"[\"password\",\"client_credentials\"]"`
}

// UpdateOAuthClientRequest 更新 OAuth client 的請求
// swagger:model UpdateOAuthClientRequest
type UpdateOAuthClientRequest struct {
	// required: true
	ClientSecret string   `json:"client_secret" validate:"required" example:"new-secret"`
	OwnerID      *int     `json:"owner_id" example:"2"`
	GrantTypes   []string `json:"grant_types" validate:"required" example:"[\"password\"]"`
}

// OAuthClientResponse 回傳給客戶端的模型
// swagger:model OAuthClientResponse
type OAuthClientResponse struct {
	ID           int       `json:"id" example:"1"`
	ClientID     string    `json:"client_id" example:"my-client"`
	ClientSecret string    `json:"client_secret" example:"secret"`
	OwnerID      *int      `json:"owner_id" example:"1"`
	GrantTypes   []string  `json:"grant_types" example:"[\"password\",\"client_credentials\"]"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateOAuthClientHandler 新增 OAuth client
// @Summary     Create a new OAuth client
// @Description 僅支援 password 與 client_credentials
// @Tags        oauth
// @Accept      json
// @Produce     json
// @Param       request body CreateOAuthClientRequest true "Create OAuth client request"
// @Success     201     {object} OAuthClientResponse
// @Failure     400     {object} dto.HTTPError
// @Failure     500     {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Router      /oauth/clients [post]
func CreateOAuthClientHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req CreateOAuthClientRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: err.Error()})
		}

		store := repository.NewOAuthClientStore(db)
		oc := &model.OAuthClient{
			ClientID:     req.ClientID,
			ClientSecret: req.ClientSecret,
			OwnerID:      req.OwnerID,
			GrantTypes:   req.GrantTypes,
		}
		if err := store.Create(c.Request().Context(), oc); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}

		return c.JSON(http.StatusCreated, OAuthClientResponse{
			ID:           oc.ID,
			ClientID:     oc.ClientID,
			ClientSecret: oc.ClientSecret,
			OwnerID:      oc.OwnerID,
			GrantTypes:   oc.GrantTypes,
			CreatedAt:    oc.CreatedAt,
			UpdatedAt:    oc.UpdatedAt,
		})
	}
}

// ListOAuthClientsHandler 列出所有 OAuth clients
// @Summary     List OAuth clients
// @Description 取得所有 OAuth clients
// @Tags        oauth
// @Produce     json
// @Success     200 {array} OAuthClientResponse
// @Failure     500 {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Router      /oauth/clients [get]
func ListOAuthClientsHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		store := repository.NewOAuthClientStore(db)
		clients, err := store.List(c.Request().Context())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}
		var resp []OAuthClientResponse
		for _, oc := range clients {
			resp = append(resp, OAuthClientResponse{
				ID:           oc.ID,
				ClientID:     oc.ClientID,
				ClientSecret: oc.ClientSecret,
				OwnerID:      oc.OwnerID,
				GrantTypes:   oc.GrantTypes,
				CreatedAt:    oc.CreatedAt,
				UpdatedAt:    oc.UpdatedAt,
			})
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
// @Success     200  {object} OAuthClientResponse
// @Failure     400  {object} dto.HTTPError
// @Failure     404  {object} dto.HTTPError
// @Failure     500  {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Router      /oauth/clients/{id} [get]
func GetOAuthClientHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid id"})
		}
		store := repository.NewOAuthClientStore(db)
		oc, err := store.GetByID(c.Request().Context(), id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return c.JSON(http.StatusNotFound, dto.HTTPError{Message: "OAuth client not found"})
			}
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}
		return c.JSON(http.StatusOK, OAuthClientResponse{
			ID:           oc.ID,
			ClientID:     oc.ClientID,
			ClientSecret: oc.ClientSecret,
			OwnerID:      oc.OwnerID,
			GrantTypes:   oc.GrantTypes,
			CreatedAt:    oc.CreatedAt,
			UpdatedAt:    oc.UpdatedAt,
		})
	}
}

// UpdateOAuthClientHandler 更新指定 ID 的 OAuth client
// @Summary     Update OAuth client
// @Tags        oauth
// @Accept      json
// @Produce     json
// @Param       id      path int true "OAuth client ID"
// @Param       request body UpdateOAuthClientRequest true "Update OAuth client request"
// @Success     200     {object} OAuthClientResponse
// @Failure     400     {object} dto.HTTPError
// @Failure     404     {object} dto.HTTPError
// @Failure     500     {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Router      /oauth/clients/{id} [put]
func UpdateOAuthClientHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid id"})
		}
		var req UpdateOAuthClientRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: err.Error()})
		}
		store := repository.NewOAuthClientStore(db)
		oc := &model.OAuthClient{
			ID:           id,
			ClientSecret: req.ClientSecret,
			OwnerID:      req.OwnerID,
			GrantTypes:   req.GrantTypes,
		}
		if err := store.Update(c.Request().Context(), oc); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return c.JSON(http.StatusNotFound, dto.HTTPError{Message: "OAuth client not found"})
			}
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}
		return c.JSON(http.StatusOK, OAuthClientResponse{
			ID:           oc.ID,
			ClientID:     oc.ClientID,
			ClientSecret: oc.ClientSecret,
			OwnerID:      oc.OwnerID,
			GrantTypes:   oc.GrantTypes,
			CreatedAt:    oc.CreatedAt,
			UpdatedAt:    oc.UpdatedAt,
		})
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
// @Router      /oauth/clients/{id} [delete]
func DeleteOAuthClientHandler(db *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid id"})
		}
		store := repository.NewOAuthClientStore(db)
		if err := store.Delete(c.Request().Context(), id); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}
		return c.NoContent(http.StatusNoContent)
	}
}
