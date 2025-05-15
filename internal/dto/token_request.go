// File: internal/dto/token_request.go
package dto

// swagger:model dto.TokenRequest
type TokenRequest struct {
	GrantType    string `form:"grant_type" validate:"required" example:"password"`
	Username     string `form:"username" example:"user@example.com"`
	Password     string `form:"password" example:"password"`
	RefreshToken string `form:"refresh_token" example:"..."`
	Scope        string `form:"scope" example:"read write"`
	ClientID     string `swaggerignore:"true"`
	ClientSecret string `swaggerignore:"true"`
}
