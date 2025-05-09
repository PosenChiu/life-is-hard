// File: internal/repository/oauth_client.go
package repository

import (
	"context"
	"fmt"

	"life-is-hard/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateOAuthClient 新增一個 OAuth Client
func CreateOAuthClient(ctx context.Context, pool *pgxpool.Pool, c *model.OAuthClient) error {
	row := pool.QueryRow(ctx, `
        INSERT INTO oauth_clients
            (client_id, client_secret, name, owner_id, redirect_uris, grant_types, scopes, is_confidential)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
        RETURNING id, created_at, updated_at
    `,
		c.ClientID,
		c.ClientSecret,
		c.Name,
		c.OwnerID,
		c.RedirectURIs,
		c.GrantTypes,
		c.Scopes,
		c.IsConfidential,
	)

	if err := row.Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return fmt.Errorf("CreateOAuthClient: %w", err)
	}
	return nil
}

// GetOAuthClientByID 以 primary key 取得 Client
func GetOAuthClientByID(ctx context.Context, pool *pgxpool.Pool, id int) (*model.OAuthClient, error) {
	c := &model.OAuthClient{}
	row := pool.QueryRow(ctx, `
        SELECT id, client_id, client_secret, name, owner_id, redirect_uris,
               grant_types, scopes, is_confidential, created_at, updated_at
        FROM oauth_clients WHERE id = $1
    `, id)

	if err := scanOAuthClient(row, c); err != nil {
		return nil, fmt.Errorf("GetOAuthClientByID: %w", err)
	}
	return c, nil
}

// GetOAuthClientByClientID 以 client_id 取得 Client
func GetOAuthClientByClientID(ctx context.Context, pool *pgxpool.Pool, clientID string) (*model.OAuthClient, error) {
	c := &model.OAuthClient{}
	row := pool.QueryRow(ctx, `
        SELECT id, client_id, client_secret, name, owner_id, redirect_uris,
               grant_types, scopes, is_confidential, created_at, updated_at
        FROM oauth_clients WHERE client_id = $1
    `, clientID)

	if err := scanOAuthClient(row, c); err != nil {
		return nil, fmt.Errorf("GetOAuthClientByClientID: %w", err)
	}
	return c, nil
}

// UpdateOAuthClient 更新 Client 相關欄位（不含 Secret）並自動更新 updated_at
func UpdateOAuthClient(ctx context.Context, pool *pgxpool.Pool, c *model.OAuthClient) error {
	_, err := pool.Exec(ctx, `
        UPDATE oauth_clients SET
            name = $1,
            owner_id = $2,
            redirect_uris = $3,
            grant_types = $4,
            scopes = $5,
            is_confidential = $6,
            updated_at = now()
        WHERE id = $7
    `,
		c.Name,
		c.OwnerID,
		c.RedirectURIs,
		c.GrantTypes,
		c.Scopes,
		c.IsConfidential,
		c.ID,
	)
	if err != nil {
		return fmt.Errorf("UpdateOAuthClient: %w", err)
	}
	return nil
}

// UpdateOAuthClientSecret 僅更新 secret
func UpdateOAuthClientSecret(ctx context.Context, pool *pgxpool.Pool, id int, newSecret string) error {
	_, err := pool.Exec(ctx, `
        UPDATE oauth_clients SET
            client_secret = $1,
            updated_at = now()
        WHERE id = $2
    `, newSecret, id)

	if err != nil {
		return fmt.Errorf("UpdateOAuthClientSecret: %w", err)
	}
	return nil
}

// DeleteOAuthClient 刪除 Client
func DeleteOAuthClient(ctx context.Context, pool *pgxpool.Pool, id int) error {
	_, err := pool.Exec(ctx, `DELETE FROM oauth_clients WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("DeleteOAuthClient: %w", err)
	}
	return nil
}

// ValidateClientCredentials 驗證 client_id / secret 是否有效
func ValidateClientCredentials(ctx context.Context, pool *pgxpool.Pool, clientID, secret string) (*model.OAuthClient, error) {
	c, err := GetOAuthClientByClientID(ctx, pool, clientID)
	if err != nil {
		return nil, err
	}
	if c == nil || !c.IsConfidential || c.ClientSecret != secret {
		return nil, nil
	}
	return c, nil
}

// internal helper
func scanOAuthClient(row pgx.Row, c *model.OAuthClient) error {
	return row.Scan(
		&c.ID,
		&c.ClientID,
		&c.ClientSecret,
		&c.Name,
		&c.OwnerID,
		&c.RedirectURIs,
		&c.GrantTypes,
		&c.Scopes,
		&c.IsConfidential,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
}
