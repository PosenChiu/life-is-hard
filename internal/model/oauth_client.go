package model

import "time"

type OAuthClient struct {
	ClientID     string    `db:"client_id" json:"client_id"`
	ClientSecret string    `db:"client_secret" json:"client_secret"`
	UserID       int       `db:"user_id" json:"user_id"`
	GrantTypes   []string  `db:"grant_types" json:"grant_types"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
