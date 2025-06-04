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

type stubValidator struct{ err error }

func (s *stubValidator) Validate(i interface{}) error { return s.err }

type fakeRow struct {
	user *model.User
	err  error
}

func (r *fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	u := r.user
	*dest[0].(*int) = u.ID
	*dest[1].(*string) = u.Name
	*dest[2].(*string) = u.Email
	*dest[3].(*string) = u.PasswordHash
	*dest[4].(*time.Time) = u.CreatedAt
	*dest[5].(*bool) = u.IsAdmin
	return nil
}

func newContext(e *echo.Echo, body string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestLoginHandler(t *testing.T) {
	e := echo.New()

	t.Run("bind error", func(t *testing.T) {
		e.Validator = &stubValidator{}
		ctx, rec := newContext(e, "{bad json")
		err := LoginHandler(&database.FakeDB{})(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "無效的表單資料")
	})

	t.Run("validate error", func(t *testing.T) {
		e.Validator = &stubValidator{err: errors.New("v")}
		ctx, rec := newContext(e, `{"username":"u","password":"p"}`)
		err := LoginHandler(&database.FakeDB{})(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "v")
	})

	t.Run("user lookup fail", func(t *testing.T) {
		e.Validator = &stubValidator{}
		db := &database.FakeDB{QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &fakeRow{err: errors.New("no rows")}
		}}
		ctx, rec := newContext(e, `{"username":"u","password":"p"}`)
		err := LoginHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("auth fail", func(t *testing.T) {
		e.Validator = &stubValidator{}
		hash, _ := service.HashPassword("good")
		sample := &model.User{ID: 1, Name: "u", Email: "e", PasswordHash: hash, CreatedAt: time.Now()}
		db := &database.FakeDB{QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &fakeRow{user: sample}
		}}
		ctx, rec := newContext(e, `{"username":"u","password":"bad"}`)
		err := LoginHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("token issue fail", func(t *testing.T) {
		e.Validator = &stubValidator{}
		hash, _ := service.HashPassword("pw")
		sample := &model.User{ID: 2, Name: "u", Email: "e", PasswordHash: hash, CreatedAt: time.Now()}
		db := &database.FakeDB{QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &fakeRow{user: sample}
		}}
		t.Setenv("JWT_SECRET", "")
		ctx, rec := newContext(e, `{"username":"u","password":"pw"}`)
		err := LoginHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
		require.Contains(t, rec.Body.String(), "failed to issue token")
	})

	t.Run("success", func(t *testing.T) {
		e.Validator = &stubValidator{}
		hash, _ := service.HashPassword("pw")
		sample := &model.User{ID: 3, Name: "u", Email: "e", PasswordHash: hash, CreatedAt: time.Now()}
		db := &database.FakeDB{QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &fakeRow{user: sample}
		}}
		t.Setenv("JWT_SECRET", "secret")
		ctx, rec := newContext(e, `{"username":"u","password":"pw"}`)
		err := LoginHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), "access_token")
	})
}
