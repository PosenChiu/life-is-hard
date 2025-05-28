// File: internal/repository/oauth_client.go

package repository

import (
	"context"
	"fmt"

	"life-is-hard/internal/database"
	"life-is-hard/internal/model"
)

func GetOAuthClientByClientID(ctx context.Context, q database.Querier, clientID string) (*model.OAuthClient, error) {
	row := q.QueryRow(ctx,
		`SELECT client_id, client_secret, user_id, grant_types, created_at, updated_at
         FROM oauth_clients
         WHERE client_id = $1`,
		clientID,
	)
	var c model.OAuthClient
	if err := row.Scan(
		&c.ClientID,
		&c.ClientSecret,
		&c.UserID,
		&c.GrantTypes,
		&c.CreatedAt,
		&c.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("GetOAuthClientByClientID: %w", err)
	}
	return &c, nil
}

func CreateOAuthClient(ctx context.Context, q database.Querier, c *model.OAuthClient) error {
	row := q.QueryRow(ctx,
		`INSERT INTO oauth_clients (client_id, client_secret, user_id, grant_types)
         VALUES ($1, $2, $3, $4)
         RETURNING client_id, created_at, updated_at`,
		c.ClientID,
		c.ClientSecret,
		c.UserID,
		c.GrantTypes,
	)
	if err := row.Scan(
		&c.ClientID,
		&c.CreatedAt,
		&c.UpdatedAt,
	); err != nil {
		return fmt.Errorf("CreateOAuthClient: %w", err)
	}
	return nil
}

func UpdateOAuthClient(ctx context.Context, q database.Querier, c *model.OAuthClient) error {
	row := q.QueryRow(ctx,
		`UPDATE oauth_clients
         SET client_secret = $1, owner_id = $2, grant_types = $3, updated_at = now()
         WHERE client_id = $4
         RETURNING updated_at`,
		c.ClientSecret,
		c.UserID,
		c.GrantTypes,
		c.ClientID,
	)
	if err := row.Scan(
		&c.UpdatedAt,
	); err != nil {
		return fmt.Errorf("UpdateOAuthClient: %w", err)
	}
	return nil
}

func DeleteOAuthClient(ctx context.Context, q database.Querier, clientID string) error {
	_, err := q.Exec(ctx,
		`DELETE FROM oauth_clients WHERE client_id = $1`,
		clientID,
	)
	if err != nil {
		return fmt.Errorf("DeleteOAuthClient: %w", err)
	}
	return nil
}

func ListOAuthClients(ctx context.Context, q database.Querier, userID int) ([]model.OAuthClient, error) {
	rows, err := q.Query(ctx,
		`SELECT client_id, client_secret, owner_id, grant_types, created_at, updated_at
         FROM oauth_clients
		 WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("ListOAuthClients: %w", err)
	}
	defer rows.Close()
	var clients []model.OAuthClient
	for rows.Next() {
		var c model.OAuthClient
		if err := rows.Scan(
			&c.ClientID,
			&c.ClientSecret,
			&c.UserID,
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
