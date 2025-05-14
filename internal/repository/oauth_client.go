// File: internal/repository/oauth_client.go

package repository

import (
	"context"
	"fmt"

	"life-is-hard/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GetOAuthClientByID finds an OAuthClient by its primary key ID.
func GetOAuthClientByID(ctx context.Context, pool *pgxpool.Pool, id int) (*model.OAuthClient, error) {
	var c model.OAuthClient
	err := pool.QueryRow(ctx,
		`SELECT id, client_id, client_secret, owner_id, grant_types, created_at, updated_at
         FROM oauth_clients
         WHERE id = $1`, id).Scan(
		&c.ID,
		&c.ClientID,
		&c.ClientSecret,
		&c.OwnerID,
		&c.GrantTypes,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetOAuthClientByID: %w", err)
	}
	return &c, nil
}

// GetOAuthClientByClientID finds an OAuthClient by its client_id.
func GetOAuthClientByClientID(ctx context.Context, pool *pgxpool.Pool, clientID string) (*model.OAuthClient, error) {
	var c model.OAuthClient
	err := pool.QueryRow(ctx,
		`SELECT id, client_id, client_secret, owner_id, grant_types, created_at, updated_at
         FROM oauth_clients
         WHERE client_id = $1`, clientID).Scan(
		&c.ID,
		&c.ClientID,
		&c.ClientSecret,
		&c.OwnerID,
		&c.GrantTypes,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetOAuthClientByClientID: %w", err)
	}
	return &c, nil
}

// CreateOAuthClient inserts a new OAuthClient and fills ID, CreatedAt, UpdatedAt.
func CreateOAuthClient(ctx context.Context, pool *pgxpool.Pool, c *model.OAuthClient) error {
	row := pool.QueryRow(ctx,
		`INSERT INTO oauth_clients (client_id, client_secret, owner_id, grant_types)
         VALUES ($1, $2, $3, $4)
         RETURNING id, created_at, updated_at`,
		c.ClientID, c.ClientSecret, c.OwnerID, c.GrantTypes,
	)
	if err := row.Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return fmt.Errorf("CreateOAuthClient: %w", err)
	}
	return nil
}

// UpdateOAuthClient updates an existing OAuthClient (excluding client_id) and fills UpdatedAt.
func UpdateOAuthClient(ctx context.Context, pool *pgxpool.Pool, c *model.OAuthClient) error {
	row := pool.QueryRow(ctx,
		`UPDATE oauth_clients
         SET client_secret = $1, owner_id = $2, grant_types = $3, updated_at = now()
         WHERE id = $4
         RETURNING updated_at`,
		c.ClientSecret, c.OwnerID, c.GrantTypes, c.ID,
	)
	if err := row.Scan(&c.UpdatedAt); err != nil {
		return fmt.Errorf("UpdateOAuthClient: %w", err)
	}
	return nil
}

// DeleteOAuthClient deletes an OAuthClient by its ID.
func DeleteOAuthClient(ctx context.Context, pool *pgxpool.Pool, id int) error {
	cmd, err := pool.Exec(ctx,
		`DELETE FROM oauth_clients WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("DeleteOAuthClient: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("DeleteOAuthClient: no row with id %d", id)
	}
	return nil
}

// ListOAuthClients returns all OAuthClient records.
func ListOAuthClients(ctx context.Context, pool *pgxpool.Pool) ([]model.OAuthClient, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, client_id, client_secret, owner_id, grant_types, created_at, updated_at
         FROM oauth_clients`,
	)
	if err != nil {
		return nil, fmt.Errorf("ListOAuthClients: %w", err)
	}
	defer rows.Close()

	var clients []model.OAuthClient
	for rows.Next() {
		var c model.OAuthClient
		if err := rows.Scan(
			&c.ID,
			&c.ClientID,
			&c.ClientSecret,
			&c.OwnerID,
			&c.GrantTypes,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan OAuthClient: %w", err)
		}
		clients = append(clients, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return clients, nil
}
