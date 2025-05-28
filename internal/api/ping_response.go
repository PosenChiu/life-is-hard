// File: internal/api/ping_response.go
package api

// swagger:model api.PingResponse
type PingResponse struct {
	Message string `json:"message" example:"pong"`
}
