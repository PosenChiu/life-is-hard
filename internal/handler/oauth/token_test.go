package oauth

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"life-is-hard/internal/cache"
	"life-is-hard/internal/database"
	"life-is-hard/internal/model"
	"life-is-hard/internal/service"
	"life-is-hard/internal/store"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

// build context
func newTokenCtx(e *echo.Echo, body, auth string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

type errBinder struct{}

func (errBinder) Bind(i any, c echo.Context) error { return errors.New("bind") }

type stubValidator struct{}

func (stubValidator) Validate(i any) error { return nil }

func restoreGlobals() {
	getOAuthClientByClientID = store.GetOAuthClientByClientID
	getUserByName = store.GetUserByName
	getUserByID = store.GetUserByID
	authenticateUser = service.AuthenticateUser
	issueAccessToken = service.IssueAccessToken
	issueRefreshToken = service.IssueRefreshToken
	issueClientAccessToken = service.IssueClientAccessToken
	validateRefreshToken = service.ValidateRefreshToken
}

func TestTokenHandler(t *testing.T) {
	defer restoreGlobals()
	e := echo.New()
	e.Validator = stubValidator{}

	// bind error
	e.Binder = errBinder{}
	ctx, rec := newTokenCtx(e, "", "")
	require.NoError(t, TokenHandler(nil, nil)(ctx))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// bad auth prefix
	e.Binder = &echo.DefaultBinder{}
	ctx, rec = newTokenCtx(e, "grant_type=password", "bad")
	require.NoError(t, TokenHandler(nil, nil)(ctx))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// bad base64
	ctx, rec = newTokenCtx(e, "grant_type=password", "Basic ???")
	require.NoError(t, TokenHandler(nil, nil)(ctx))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// missing colon
	ctx, rec = newTokenCtx(e, "grant_type=password", "Basic "+base64.StdEncoding.EncodeToString([]byte("abc")))
	require.NoError(t, TokenHandler(nil, nil)(ctx))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// invalid client credentials
	auth := "Basic " + base64.StdEncoding.EncodeToString([]byte("id:sec"))
	getOAuthClientByClientID = func(_ context.Context, _ database.DB, _ string) (*model.OAuthClient, error) {
		return &model.OAuthClient{ClientID: "id", ClientSecret: "other"}, nil
	}
	ctx, rec = newTokenCtx(e, "grant_type=password", auth)
	require.NoError(t, TokenHandler(nil, nil)(ctx))
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	// unauthorized grant type
	getOAuthClientByClientID = func(_ context.Context, _ database.DB, _ string) (*model.OAuthClient, error) {
		return &model.OAuthClient{ClientID: "id", ClientSecret: "sec", GrantTypes: []string{"client_credentials"}}, nil
	}
	ctx, rec = newTokenCtx(e, "grant_type=password", auth)
	require.NoError(t, TokenHandler(nil, nil)(ctx))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// password grant user not found
	getOAuthClientByClientID = func(_ context.Context, _ database.DB, _ string) (*model.OAuthClient, error) {
		return &model.OAuthClient{ClientID: "id", ClientSecret: "sec", GrantTypes: []string{"password"}}, nil
	}
	getUserByName = func(context.Context, database.DB, string) (*model.User, error) { return nil, errors.New("no") }
	ctx, rec = newTokenCtx(e, "grant_type=password", auth)
	require.NoError(t, TokenHandler(nil, nil)(ctx))
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	// password grant auth fail
	getUserByName = func(context.Context, database.DB, string) (*model.User, error) { return &model.User{}, nil }
	authenticateUser = func(context.Context, model.User, string) error { return errors.New("bad") }
	ctx, rec = newTokenCtx(e, "grant_type=password", auth)
	require.NoError(t, TokenHandler(nil, nil)(ctx))
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	// password grant issueAccessToken error
	authenticateUser = func(context.Context, model.User, string) error { return nil }
	issueAccessToken = func(model.User, time.Duration) (string, error) { return "", errors.New("x") }
	ctx, rec = newTokenCtx(e, "grant_type=password", auth)
	require.NoError(t, TokenHandler(nil, nil)(ctx))
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	// password grant issueRefreshToken error
	issueAccessToken = func(model.User, time.Duration) (string, error) { return "tok", nil }
	issueRefreshToken = func(context.Context, cache.Cache, int, string, bool, time.Duration) (string, error) {
		return "", errors.New("x")
	}
	ctx, rec = newTokenCtx(e, "grant_type=password", auth)
	require.NoError(t, TokenHandler(nil, nil)(ctx))
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	// password grant success
	issueRefreshToken = func(context.Context, cache.Cache, int, string, bool, time.Duration) (string, error) { return "rt", nil }
	ctx, rec = newTokenCtx(e, "grant_type=password", auth)
	require.NoError(t, TokenHandler(nil, nil)(ctx))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "rt")

	// client_credentials owner error
	getOAuthClientByClientID = func(_ context.Context, _ database.DB, _ string) (*model.OAuthClient, error) {
		return &model.OAuthClient{ClientID: "id", ClientSecret: "sec", GrantTypes: []string{"client_credentials"}, UserID: 1}, nil
	}
	getUserByID = func(context.Context, database.DB, int) (*model.User, error) { return nil, errors.New("x") }
	ctx, rec = newTokenCtx(e, "grant_type=client_credentials", auth)
	require.NoError(t, TokenHandler(nil, nil)(ctx))
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	// client_credentials issue error
	getUserByID = func(context.Context, database.DB, int) (*model.User, error) { return &model.User{}, nil }
	issueClientAccessToken = func(model.User, model.OAuthClient, time.Duration) (string, error) { return "", errors.New("x") }
	ctx, rec = newTokenCtx(e, "grant_type=client_credentials", auth)
	require.NoError(t, TokenHandler(nil, nil)(ctx))
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	// client_credentials success
	issueClientAccessToken = func(model.User, model.OAuthClient, time.Duration) (string, error) { return "tok", nil }
	ctx, rec = newTokenCtx(e, "grant_type=client_credentials", auth)
	require.NoError(t, TokenHandler(nil, nil)(ctx))
	require.Equal(t, http.StatusOK, rec.Code)

	// refresh_token validate error
	getOAuthClientByClientID = func(_ context.Context, _ database.DB, _ string) (*model.OAuthClient, error) {
		return &model.OAuthClient{ClientID: "id", ClientSecret: "sec", GrantTypes: []string{"refresh_token"}}, nil
	}
	validateRefreshToken = func(context.Context, cache.Cache, string) (*service.RefreshTokenData, error) {
		return nil, errors.New("bad")
	}
	ctx, rec = newTokenCtx(e, "grant_type=refresh_token&refresh_token=rt", auth)
	require.NoError(t, TokenHandler(nil, nil)(ctx))
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	// refresh_token issue access error
	validateRefreshToken = func(context.Context, cache.Cache, string) (*service.RefreshTokenData, error) {
		return &service.RefreshTokenData{UserID: 2}, nil
	}
	issueAccessToken = func(model.User, time.Duration) (string, error) { return "", errors.New("x") }
	ctx, rec = newTokenCtx(e, "grant_type=refresh_token&refresh_token=rt", auth)
	require.NoError(t, TokenHandler(nil, nil)(ctx))
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	// refresh_token success
	issueAccessToken = func(model.User, time.Duration) (string, error) { return "tok", nil }
	ctx, rec = newTokenCtx(e, "grant_type=refresh_token&refresh_token=rt", auth)
	require.NoError(t, TokenHandler(nil, nil)(ctx))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "refresh_token\":\"rt")

	// unsupported grant
	getOAuthClientByClientID = func(_ context.Context, _ database.DB, _ string) (*model.OAuthClient, error) {
		return &model.OAuthClient{ClientID: "id", ClientSecret: "sec", GrantTypes: []string{"unknown"}}, nil
	}
	ctx, rec = newTokenCtx(e, "grant_type=unknown", auth)
	require.NoError(t, TokenHandler(nil, nil)(ctx))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
