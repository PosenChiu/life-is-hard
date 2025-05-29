package api

// swagger:model api.LoginRequest
type LoginRequest struct {
	Username string `form:"username" validate:"required" example:"alice"`
	Password string `form:"password" validate:"required" example:"Secret123!"`
}
