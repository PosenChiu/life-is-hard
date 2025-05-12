// File: internal/repository/oauth_client.go
package repository

import (
	"context"
	"fmt"

	"life-is-hard/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

// OAuthClientStore 負責對 oauth_clients 表的所有操作
type OAuthClientStore struct {
	pool *pgxpool.Pool
}

// NewOAuthClientStore 建立一個新的 OAuthClientStore
func NewOAuthClientStore(pool *pgxpool.Pool) *OAuthClientStore {
	return &OAuthClientStore{pool: pool}
}

// GetByID 根據主鍵 ID 查找 OAuthClient
func (s *OAuthClientStore) GetByID(ctx context.Context, id int) (*model.OAuthClient, error) {
	var c model.OAuthClient
	err := s.pool.QueryRow(ctx, `
		SELECT id, client_id, client_secret, owner_id, grant_types, created_at, updated_at
		FROM oauth_clients
		WHERE id = $1
	`, id).Scan(
		&c.ID,
		&c.ClientID,
		&c.ClientSecret,
		&c.OwnerID,
		&c.GrantTypes,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	return &c, nil
}

// GetByClientID 根據 client_id 查找 OAuthClient
func (s *OAuthClientStore) GetByClientID(ctx context.Context, clientID string) (*model.OAuthClient, error) {
	var c model.OAuthClient
	err := s.pool.QueryRow(ctx, `
		SELECT id, client_id, client_secret, owner_id, grant_types, created_at, updated_at
		FROM oauth_clients
		WHERE client_id = $1
	`, clientID).Scan(
		&c.ID,
		&c.ClientID,
		&c.ClientSecret,
		&c.OwnerID,
		&c.GrantTypes,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetByClientID: %w", err)
	}
	return &c, nil
}

// Create 插入一筆新的 OAuthClient，並回填 ID、CreatedAt、UpdatedAt
func (s *OAuthClientStore) Create(ctx context.Context, c *model.OAuthClient) error {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO oauth_clients (client_id, client_secret, owner_id, grant_types)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`, c.ClientID, c.ClientSecret, c.OwnerID, c.GrantTypes)

	if err := row.Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return fmt.Errorf("Create: %w", err)
	}
	return nil
}

// Update 更新既有的 OAuthClient（除 ClientID 之外的欄位），並回填 UpdatedAt
func (s *OAuthClientStore) Update(ctx context.Context, c *model.OAuthClient) error {
	row := s.pool.QueryRow(ctx, `
		UPDATE oauth_clients
		SET client_secret = $1, owner_id = $2, grant_types = $3, updated_at = now()
		WHERE id = $4
		RETURNING updated_at
	`, c.ClientSecret, c.OwnerID, c.GrantTypes, c.ID)

	if err := row.Scan(&c.UpdatedAt); err != nil {
		return fmt.Errorf("Update: %w", err)
	}
	return nil
}

// Delete 刪除指定 ID 的 OAuthClient
func (s *OAuthClientStore) Delete(ctx context.Context, id int) error {
	cmd, err := s.pool.Exec(ctx, `
		DELETE FROM oauth_clients WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("Delete: no row with id %d", id)
	}
	return nil
}
