// File: internal/model/user.go
package model

import "time"

type User struct {
	ID           int       `db:"id" json:"id"`
	Name         string    `db:"name" json:"name"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"password_hash"`
	IsAdmin      bool      `db:"is_admin" json:"is_admin"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
