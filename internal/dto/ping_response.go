// File: internal/dto/ping_response.go
package dto

// swagger:model dto.PingResponse
type PingResponse struct {
	Message string `json:"message" example:"pong"`
}
