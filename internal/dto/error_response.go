// File: internal/dto/error_response.go
package dto

// swagger:model dto.HTTPError
type ErrorResponse struct {
	Message string `json:"message"`
}
