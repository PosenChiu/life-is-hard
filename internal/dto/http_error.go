// File: internal/dto/http_error.go
package dto

// HTTPError 全域錯誤響應模型
// swagger:model dto.HTTPError
type HTTPError struct {
	// message 錯誤描述
	Message string `json:"message"`
}
