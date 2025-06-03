package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"life-is-hard/internal/database"
	"life-is-hard/internal/model"
	"life-is-hard/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

// helper to build echo context
func newLoginCtx(e *echo.Echo, body string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

type errBinder struct{}

func (errBinder) Bind(i any, c echo.Context) error { return errors.New("bind") }

type errValidator struct{}

func (errValidator) Validate(i any) error { return errors.New("v") }

type okValidator struct{}

func (okValidator) Validate(i any) error { return nil }

type fakeRow struct {
	u   model.User
	err error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*dest[0].(*int) = r.u.ID
	*dest[1].(*string) = r.u.Name
	*dest[2].(*string) = r.u.Email
	*dest[3].(*string) = r.u.PasswordHash
	*dest[4].(*time.Time) = r.u.CreatedAt
	*dest[5].(*bool) = r.u.IsAdmin
	return nil
}

func TestLoginHandler(t *testing.T) {

	// bind error
	e := echo.New()
	e.Binder = errBinder{}
	ctx, rec := newLoginCtx(e, "")
	h := LoginHandler(&database.FakeDB{})
	require.NoError(t, h(ctx))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// validate error
	e = echo.New()
	e.Validator = errValidator{}
	ctx, rec = newLoginCtx(e, "username=a&password=b")
	h = LoginHandler(&database.FakeDB{QueryRowFn: func(context.Context, string, ...any) pgx.Row { return fakeRow{} }})
	require.NoError(t, h(ctx))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// user not found
	e = echo.New()
	e.Validator = okValidator{}
	ctx, rec = newLoginCtx(e, "username=a&password=b")
	h = LoginHandler(&database.FakeDB{QueryRowFn: func(context.Context, string, ...any) pgx.Row { return fakeRow{err: errors.New("no")} }})
	require.NoError(t, h(ctx))
	t.Log("notfound", rec.Body.String())
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	// authenticate error
	e = echo.New()
	e.Validator = okValidator{}
	ctx, rec = newLoginCtx(e, "username=a&password=b")
	badHash, _ := service.HashPassword("other")
	h = LoginHandler(&database.FakeDB{QueryRowFn: func(context.Context, string, ...any) pgx.Row { return fakeRow{u: model.User{PasswordHash: badHash}} }})
	require.NoError(t, h(ctx))
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	t.Log("auth err", rec.Body.String())

	// issue token error (JWT_SECRET not set)
	e = echo.New()
	e.Validator = okValidator{}
	ctx, rec = newLoginCtx(e, "username=a&password=b")
	goodHash, _ := service.HashPassword("b")
	h = LoginHandler(&database.FakeDB{QueryRowFn: func(context.Context, string, ...any) pgx.Row { return fakeRow{u: model.User{PasswordHash: goodHash}} }})
	require.NoError(t, h(ctx))
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	// success
	e = echo.New()
	e.Validator = okValidator{}
	ctx, rec = newLoginCtx(e, "username=a&password=b")
	t.Setenv("JWT_SECRET", "s")
	h = LoginHandler(&database.FakeDB{QueryRowFn: func(context.Context, string, ...any) pgx.Row {
		return fakeRow{u: model.User{ID: 1, PasswordHash: goodHash}}
	}})
	require.NoError(t, h(ctx))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "access_token")
}
