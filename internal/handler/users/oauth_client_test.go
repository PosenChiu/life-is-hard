package users

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"life-is-hard/internal/database"
	"life-is-hard/internal/middleware"
	"life-is-hard/internal/model"
	"life-is-hard/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

// fakeRow implements pgx.Row for OAuth client queries
type fakeRow struct {
	scanErr error
	client  *model.OAuthClient
}

func (r *fakeRow) Scan(dest ...any) error {
	if r.scanErr != nil {
		return r.scanErr
	}
	c := r.client
	switch len(dest) {
	case 6:
		*dest[0].(*string) = c.ClientID
		*dest[1].(*string) = c.ClientSecret
		*dest[2].(*int) = c.UserID
		*dest[3].(*[]string) = c.GrantTypes
		*dest[4].(*time.Time) = c.CreatedAt
		*dest[5].(*time.Time) = c.UpdatedAt
	case 3:
		*dest[0].(*string) = c.ClientID
		*dest[1].(*time.Time) = c.CreatedAt
		*dest[2].(*time.Time) = c.UpdatedAt
	case 1:
		*dest[0].(*time.Time) = c.UpdatedAt
	default:
		panic("unexpected dest count")
	}
	return nil
}

// fakeRows implements pgx.Rows for listing clients
type fakeRows struct {
	data    []model.OAuthClient
	idx     int
	scanErr error
	err     error
}

func (r *fakeRows) Close()                                       {}
func (r *fakeRows) Err() error                                   { return r.err }
func (r *fakeRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (r *fakeRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *fakeRows) Next() bool {
	ok := r.idx < len(r.data)
	if ok {
		r.idx++
	}
	return ok
}
func (r *fakeRows) Scan(dest ...any) error {
	if r.scanErr != nil {
		return r.scanErr
	}
	c := r.data[r.idx-1]
	*dest[0].(*string) = c.ClientID
	*dest[1].(*string) = c.ClientSecret
	*dest[2].(*int) = c.UserID
	*dest[3].(*[]string) = c.GrantTypes
	*dest[4].(*time.Time) = c.CreatedAt
	*dest[5].(*time.Time) = c.UpdatedAt
	return nil
}
func (r *fakeRows) Values() ([]any, error) { return nil, nil }
func (r *fakeRows) RawValues() [][]byte    { return nil }
func (r *fakeRows) Conn() *pgx.Conn        { return nil }

// helpers for creating echo contexts
func newJSONCtx(e *echo.Echo, method, path, body string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func newClientCtx(e *echo.Echo, method, id, body string) (echo.Context, *httptest.ResponseRecorder) {
	c, rec := newJSONCtx(e, method, "/users/me/oauth-clients/"+id, body)
	c.SetPath("/users/me/oauth-clients/:client_id")
	c.SetParamNames("client_id")
	c.SetParamValues(id)
	return c, rec
}

// sample OAuth client for tests
var sampleClient = model.OAuthClient{
	ClientID:     "cid",
	ClientSecret: "sec",
	UserID:       1,
	GrantTypes:   []string{"password"},
	CreatedAt:    time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	UpdatedAt:    time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
}

func TestCreateMyOAuthClientHandler(t *testing.T) {
	e := echo.New()
	e.Validator = &stubValidator{}

	t.Run("no claims", func(t *testing.T) {
		ctx, rec := newJSONCtx(e, http.MethodPost, "/users/me/oauth-clients", `{}`)
		err := CreateMyOAuthClientHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("bind error", func(t *testing.T) {
		ctx, rec := newJSONCtx(e, http.MethodPost, "/users/me/oauth-clients", `{bad`)
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := CreateMyOAuthClientHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("validate error", func(t *testing.T) {
		e.Validator = &stubValidator{err: errors.New("v")}
		ctx, rec := newJSONCtx(e, http.MethodPost, "/users/me/oauth-clients", `{"client_id":"c"}`)
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := CreateMyOAuthClientHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "v")
		e.Validator = &stubValidator{}
	})

	t.Run("store error", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(context.Context, string, ...any) pgx.Row {
			return &fakeRow{scanErr: errors.New("fail")}
		}}
		body := `{"client_id":"c","client_secret":"s","grant_types":["password"]}`
		ctx, rec := newJSONCtx(e, http.MethodPost, "/users/me/oauth-clients", body)
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 2})
		err := CreateMyOAuthClientHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(context.Context, string, ...any) pgx.Row {
			c := sampleClient
			c.ClientID = "new"
			return &fakeRow{client: &c}
		}}
		body := `{"client_id":"new","client_secret":"s","grant_types":["password"]}`
		ctx, rec := newJSONCtx(e, http.MethodPost, "/users/me/oauth-clients", body)
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := CreateMyOAuthClientHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, rec.Code)
		require.Contains(t, rec.Body.String(), "\"client_id\":\"new\"")
	})
}

func TestListMyOAuthClientsHandler(t *testing.T) {
	e := echo.New()

	t.Run("no claims", func(t *testing.T) {
		ctx, rec := newJSONCtx(e, http.MethodGet, "/users/me/oauth-clients", "")
		err := ListMyOAuthClientsHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("store error", func(t *testing.T) {
		db := &database.FakeDB{QueryFn: func(context.Context, string, ...any) (pgx.Rows, error) {
			return nil, errors.New("db")
		}}
		ctx, rec := newJSONCtx(e, http.MethodGet, "/users/me/oauth-clients", "")
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := ListMyOAuthClientsHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		rows := &fakeRows{data: []model.OAuthClient{sampleClient, sampleClient}}
		db := &database.FakeDB{QueryFn: func(context.Context, string, ...any) (pgx.Rows, error) { return rows, nil }}
		ctx, rec := newJSONCtx(e, http.MethodGet, "/users/me/oauth-clients", "")
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := ListMyOAuthClientsHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), "client_secret")
	})
}

func TestGetMyOAuthClientHandler(t *testing.T) {
	e := echo.New()

	t.Run("no claims", func(t *testing.T) {
		ctx, rec := newClientCtx(e, http.MethodGet, "cid", "")
		err := GetMyOAuthClientHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("store error", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(context.Context, string, ...any) pgx.Row {
			return &fakeRow{scanErr: errors.New("fail")}
		}}
		ctx, rec := newClientCtx(e, http.MethodGet, "cid", "")
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := GetMyOAuthClientHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("not owner", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(context.Context, string, ...any) pgx.Row {
			c := sampleClient
			c.UserID = 2
			return &fakeRow{client: &c}
		}}
		ctx, rec := newClientCtx(e, http.MethodGet, "cid", "")
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := GetMyOAuthClientHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(context.Context, string, ...any) pgx.Row {
			return &fakeRow{client: &sampleClient}
		}}
		ctx, rec := newClientCtx(e, http.MethodGet, "cid", "")
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := GetMyOAuthClientHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), "\"client_id\":\"cid\"")
	})
}

func TestUpdateMyOAuthClientHandler(t *testing.T) {
	e := echo.New()
	e.Validator = &stubValidator{}

	body := `{"client_secret":"ns","grant_types":["password"]}`

	t.Run("no claims", func(t *testing.T) {
		ctx, rec := newClientCtx(e, http.MethodPut, "cid", body)
		err := UpdateMyOAuthClientHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("bind error", func(t *testing.T) {
		ctx, rec := newClientCtx(e, http.MethodPut, "cid", `{bad`)
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := UpdateMyOAuthClientHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("validate error", func(t *testing.T) {
		e.Validator = &stubValidator{err: errors.New("v")}
		ctx, rec := newClientCtx(e, http.MethodPut, "cid", body)
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := UpdateMyOAuthClientHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		e.Validator = &stubValidator{}
	})

	t.Run("get error", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(context.Context, string, ...any) pgx.Row {
			return &fakeRow{scanErr: errors.New("fail")}
		}}
		ctx, rec := newClientCtx(e, http.MethodPut, "cid", body)
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := UpdateMyOAuthClientHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("not owner", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(context.Context, string, ...any) pgx.Row {
			c := sampleClient
			c.UserID = 2
			return &fakeRow{client: &c}
		}}
		ctx, rec := newClientCtx(e, http.MethodPut, "cid", body)
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := UpdateMyOAuthClientHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("update error", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(_ context.Context, q string, _ ...any) pgx.Row {
			if strings.HasPrefix(q, "UPDATE") {
				return &fakeRow{scanErr: errors.New("up")}
			}
			return &fakeRow{client: &sampleClient}
		}}
		ctx, rec := newClientCtx(e, http.MethodPut, "cid", body)
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := UpdateMyOAuthClientHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		updated := sampleClient
		updated.ClientSecret = "ns"
		updated.UpdatedAt = updated.UpdatedAt.Add(time.Hour)
		db := &database.FakeDB{QueryRowFn: func(context.Context, string, ...any) pgx.Row {
			return &fakeRow{client: &updated}
		}}
		ctx, rec := newClientCtx(e, http.MethodPut, "cid", body)
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := UpdateMyOAuthClientHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), "\"client_secret\":\"ns\"")
	})
}

func TestDeleteMyOAuthClientHandler(t *testing.T) {
	e := echo.New()

	t.Run("no claims", func(t *testing.T) {
		ctx, rec := newClientCtx(e, http.MethodDelete, "cid", "")
		err := DeleteMyOAuthClientHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("get error", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(context.Context, string, ...any) pgx.Row { return &fakeRow{scanErr: errors.New("fail")} }}
		ctx, rec := newClientCtx(e, http.MethodDelete, "cid", "")
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := DeleteMyOAuthClientHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("not owner", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(context.Context, string, ...any) pgx.Row {
			c := sampleClient
			c.UserID = 2
			return &fakeRow{client: &c}
		}}
		ctx, rec := newClientCtx(e, http.MethodDelete, "cid", "")
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := DeleteMyOAuthClientHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("delete error", func(t *testing.T) {
		db := &database.FakeDB{
			QueryRowFn: func(context.Context, string, ...any) pgx.Row { return &fakeRow{client: &sampleClient} },
			ExecFn: func(context.Context, string, ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, errors.New("del")
			},
		}
		ctx, rec := newClientCtx(e, http.MethodDelete, "cid", "")
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := DeleteMyOAuthClientHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		db := &database.FakeDB{
			QueryRowFn: func(context.Context, string, ...any) pgx.Row { return &fakeRow{client: &sampleClient} },
			ExecFn:     func(context.Context, string, ...any) (pgconn.CommandTag, error) { return pgconn.CommandTag{}, nil },
		}
		ctx, rec := newClientCtx(e, http.MethodDelete, "cid", "")
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := DeleteMyOAuthClientHandler(db)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, rec.Code)
	})
}
