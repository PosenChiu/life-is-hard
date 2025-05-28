// File: internal/api/login_response.go
package api

// swagger:model api.LoginResponse
type LoginResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOi..."`
}
