package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"life-is-hard/internal/model"
	"life-is-hard/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func newContext(auth string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestExtractClaims(t *testing.T) {
	t.Setenv("JWT_SECRET", "testsecret")

	// missing header
	ctx, _ := newContext("")
	_, err := extractClaims(ctx)
	require.Error(t, err)

	// bad format
	ctx, _ = newContext("BadHeader")
	_, err = extractClaims(ctx)
	require.Error(t, err)

	// invalid token
	ctx, _ = newContext("Bearer invalid")
	_, err = extractClaims(ctx)
	require.Error(t, err)

	// valid token
	tok, err := service.IssueAccessToken(model.User{ID: 1, IsAdmin: true}, time.Minute)
	require.NoError(t, err)
	ctx, _ = newContext("Bearer " + tok)
	claims, err := extractClaims(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, claims.UserID)
	require.True(t, claims.IsAdmin)
}

func TestRequireAuth(t *testing.T) {
	t.Setenv("JWT_SECRET", "secret")
	tok, err := service.IssueAccessToken(model.User{ID: 2}, time.Minute)
	require.NoError(t, err)

	// success path
	ctx, rec := newContext("Bearer " + tok)
	called := false
	handler := RequireAuth(func(c echo.Context) error {
		called = true
		cl := c.Get(ContextUserKey).(*service.CustomClaims)
		require.Equal(t, 2, cl.UserID)
		return c.String(http.StatusOK, "ok")
	})
	require.NoError(t, handler(ctx))
	require.True(t, called)
	require.Equal(t, http.StatusOK, rec.Code)

	// missing token
	ctx, _ = newContext("")
	called = false
	err = RequireAuth(func(echo.Context) error { called = true; return nil })(ctx)
	require.Error(t, err)
	require.False(t, called)
}

func TestRequireAdmin(t *testing.T) {
	t.Setenv("JWT_SECRET", "adminsecret")
	adminTok, err := service.IssueAccessToken(model.User{ID: 3, IsAdmin: true}, time.Minute)
	require.NoError(t, err)
	userTok, err := service.IssueAccessToken(model.User{ID: 4, IsAdmin: false}, time.Minute)
	require.NoError(t, err)

	// admin ok
	ctx, rec := newContext("Bearer " + adminTok)
	called := false
	err = RequireAdmin(func(c echo.Context) error { called = true; return c.String(http.StatusOK, "admin") })(ctx)
	require.NoError(t, err)
	require.True(t, called)
	require.Equal(t, http.StatusOK, rec.Code)

	// non-admin should fail
	ctx, _ = newContext("Bearer " + userTok)
	called = false
	err = RequireAdmin(func(c echo.Context) error { called = true; return nil })(ctx)
	require.Error(t, err)
	require.False(t, called)
}
