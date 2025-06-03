package router

import (
	"net/http"
	"testing"

	"life-is-hard/internal/cache"
	"life-is-hard/internal/database"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestSetupRoutes(t *testing.T) {
	e := echo.New()
	Setup(e, &database.FakeDB{}, &cache.FakeCache{})

	got := map[string]struct{}{}
	for _, r := range e.Routes() {
		got[r.Method+" "+r.Path] = struct{}{}
	}

	expected := []string{
		http.MethodGet + " /api/ping",
		http.MethodPost + " /api/auth/login",
		http.MethodPost + " /api/oauth/token",
		http.MethodPost + " /api/users",
		http.MethodGet + " /api/users/:id",
		http.MethodPut + " /api/users/:id",
		http.MethodDelete + " /api/users/:id",
		http.MethodGet + " /api/users/me",
		http.MethodPut + " /api/users/me",
		http.MethodDelete + " /api/users/me",
		http.MethodPatch + " /api/users/me/password",
		http.MethodPost + " /api/users/me/oauth-clients",
		http.MethodGet + " /api/users/me/oauth-clients",
		http.MethodGet + " /api/users/me/oauth-clients/:client_id",
		http.MethodPut + " /api/users/me/oauth-clients/:client_id",
		http.MethodDelete + " /api/users/me/oauth-clients/:client_id",
	}

	require.Equal(t, len(expected), len(got))
	for _, k := range expected {
		_, ok := got[k]
		require.True(t, ok, "missing route %s", k)
	}
}
