// File: internal/repository/user.go
package repository

import (
	"context"
	"fmt"

	"life-is-hard/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetUserByID(ctx context.Context, pool *pgxpool.Pool, userID int) (*model.User, error) {
	row := pool.QueryRow(ctx,
		`SELECT id, name, email, password_hash, created_at, is_admin
		 FROM users WHERE id = $1`,
		userID,
	)
	u := &model.User{}
	if err := row.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.IsAdmin,
	); err != nil {
		return nil, fmt.Errorf("GetUserByID: %w", err)
	}
	return u, nil
}

func GetUserByName(ctx context.Context, pool *pgxpool.Pool, userName string) (*model.User, error) {
	row := pool.QueryRow(ctx,
		`SELECT id, name, email, password_hash, created_at, is_admin
		 FROM users WHERE name = $1`,
		userName,
	)
	u := &model.User{}
	if err := row.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.IsAdmin,
	); err != nil {
		return nil, fmt.Errorf("GetUserByName: %w", err)
	}
	return u, nil
}

func CreateUser(ctx context.Context, pool *pgxpool.Pool, u *model.User) (*model.User, error) {
	row := pool.QueryRow(ctx,
		`INSERT INTO users (name, email, password_hash, is_admin)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		u.Name,
		u.Email,
		u.PasswordHash,
		u.IsAdmin,
	)
	if err := row.Scan(&u.ID, &u.CreatedAt); err != nil {
		return nil, fmt.Errorf("CreateUser: %w", err)
	}
	return u, nil
}

func UpdateUser(ctx context.Context, pool *pgxpool.Pool, u *model.User) error {
	_, err := pool.Exec(ctx,
		`UPDATE users SET name = $1, email = $2, is_admin = $3
		 WHERE id = $4`,
		u.Name,
		u.Email,
		u.IsAdmin,
		u.ID,
	)
	if err != nil {
		return fmt.Errorf("UpdateUser: %w", err)
	}
	return nil
}

func UpdateUserPassword(ctx context.Context, pool *pgxpool.Pool, userID int, passwordHash string) error {
	_, err := pool.Exec(ctx,
		`UPDATE users
		 SET password_hash = $1
		 WHERE id = $2`,
		passwordHash,
		userID,
	)
	if err != nil {
		return fmt.Errorf("UpdateUserPassword: %w", err)
	}
	return nil
}

func DeleteUser(ctx context.Context, pool *pgxpool.Pool, ID int) error {
	_, err := pool.Exec(ctx,
		`DELETE FROM users WHERE id = $1`,
		ID,
	)
	if err != nil {
		return fmt.Errorf("DeleteUser: %w", err)
	}
	return nil
}
