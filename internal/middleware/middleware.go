package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"life-is-hard/internal/service"

	"github.com/labstack/echo/v4"
)

const ContextUserKey = "user"

func extractClaims(c echo.Context) (*service.CustomClaims, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "missing token")
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
	}
	tokenString := parts[1]
	claims, err := service.VerifyAccessToken(tokenString)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("invalid token: %v", err))
	}
	return claims, nil
}

func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, err := extractClaims(c)
		if err != nil {
			return err
		}
		c.Set(ContextUserKey, claims)
		return next(c)
	}
}

func RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return RequireAuth(func(c echo.Context) error {
		claims := c.Get(ContextUserKey).(*service.CustomClaims)
		if !claims.IsAdmin {
			return echo.NewHTTPError(http.StatusForbidden, "admin privileges required")
		}
		return next(c)
	})
}
