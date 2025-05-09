// File: internal/model/oauth_client.go
package model

import "time"

type OAuthClient struct {
	ID             int       `db:"id" json:"-"`
	ClientID       string    `db:"client_id" json:"client_id"`
	ClientSecret   string    `db:"client_secret" json:"-"`
	Name           string    `db:"name" json:"name"`
	OwnerID        *int      `db:"owner_id" json:"owner_id,omitempty"`
	RedirectURIs   []string  `db:"redirect_uris" json:"redirect_uris"`
	GrantTypes     []string  `db:"grant_types" json:"grant_types"`
	Scopes         []string  `db:"scopes" json:"scopes"`
	IsConfidential bool      `db:"is_confidential" json:"is_confidential"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}
