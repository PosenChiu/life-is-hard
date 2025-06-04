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
	"life-is-hard/internal/store"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type stubValidator struct{ err error }

func (s *stubValidator) Validate(i interface{}) error { return s.err }

func newFormCtx(e *echo.Echo, body string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func newParamCtx(e *echo.Echo, val string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodGet, "/users/"+val, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:user_id")
	c.SetParamNames("user_id")
	c.SetParamValues(val)
	return c, rec
}

func newUpdateCtx(e *echo.Echo, id, body string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPut, "/users/"+id, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(id)
	return c, rec
}

func newMeCtx(e *echo.Echo, method, body string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, "/users/me", strings.NewReader(body))
	if body != "" {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func restore() {
	hashPassword = service.HashPassword
	authenticateUser = service.AuthenticateUser
	createUser = store.CreateUser
	getUserByID = store.GetUserByID
	updateUser = store.UpdateUser
	updateUserPassword = store.UpdateUserPassword
	deleteUser = store.DeleteUser
}

func TestCreateUserHandler(t *testing.T) {
	e := echo.New()

	t.Run("bind error", func(t *testing.T) {
		t.Cleanup(restore)
		e.Validator = &stubValidator{}
		ctx, rec := newFormCtx(e, "%")
		err := CreateUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "invalid form data")
	})

	t.Run("validate error", func(t *testing.T) {
		t.Cleanup(restore)
		e.Validator = &stubValidator{err: errors.New("v")}
		ctx, rec := newFormCtx(e, "name=a&email=a@b.com&password=p&is_admin=true")
		err := CreateUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "v")
	})

	t.Run("hash error", func(t *testing.T) {
		t.Cleanup(restore)
		e.Validator = &stubValidator{}
		hashPassword = func(string) (string, error) { return "", errors.New("hash") }
		ctx, rec := newFormCtx(e, "name=a&email=a@b.com&password=p&is_admin=true")
		err := CreateUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "failed to hash password")
	})

	t.Run("bad email", func(t *testing.T) {
		t.Cleanup(restore)
		e.Validator = &stubValidator{}
		hashPassword = func(string) (string, error) { return "h", nil }
		ctx, rec := newFormCtx(e, "name=a&email=bad&password=p&is_admin=true")
		err := CreateUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "invalid email format")
	})

	t.Run("create error", func(t *testing.T) {
		t.Cleanup(restore)
		e.Validator = &stubValidator{}
		hashPassword = func(string) (string, error) { return "h", nil }
		createUser = func(_ context.Context, _ database.DB, u *model.User) (*model.User, error) {
			return nil, errors.New("c")
		}
		ctx, rec := newFormCtx(e, "name=a&email=a@b.com&password=p&is_admin=true")
		err := CreateUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Cleanup(restore)
		e.Validator = &stubValidator{}
		now := time.Now().UTC()
		hashPassword = func(p string) (string, error) { require.Equal(t, "p", p); return "h", nil }
		var gotEmail string
		createUser = func(_ context.Context, _ database.DB, u *model.User) (*model.User, error) {
			gotEmail = u.Email
			u.ID = 1
			u.CreatedAt = now
			return u, nil
		}
		ctx, rec := newFormCtx(e, "name=A&email=Alice@EXAMPLE.com&password=p&is_admin=true")
		err := CreateUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, rec.Code)
		require.Equal(t, "alice@example.com", gotEmail)
		require.Contains(t, rec.Body.String(), "\"id\":1")
	})
}

func TestGetUserHandler(t *testing.T) {
	e := echo.New()
	t.Run("bad id", func(t *testing.T) {
		t.Cleanup(restore)
		ctx, rec := newParamCtx(e, "x")
		err := GetUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		t.Cleanup(restore)
		getUserByID = func(context.Context, database.DB, int) (*model.User, error) { return nil, errors.New("no") }
		ctx, rec := newParamCtx(e, "1")
		err := GetUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Cleanup(restore)
		now := time.Now().UTC()
		getUserByID = func(context.Context, database.DB, int) (*model.User, error) {
			return &model.User{ID: 1, Name: "n", Email: "e", CreatedAt: now}, nil
		}
		ctx, rec := newParamCtx(e, "1")
		err := GetUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), "\"id\":1")
	})
}

func TestUpdateUserHandler(t *testing.T) {
	e := echo.New()
	e.Validator = &stubValidator{}

	t.Run("bad id", func(t *testing.T) {
		t.Cleanup(restore)
		ctx, rec := newUpdateCtx(e, "x", "")
		err := UpdateUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("bind error", func(t *testing.T) {
		t.Cleanup(restore)
		ctx, rec := newUpdateCtx(e, "1", "%")
		err := UpdateUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("validate error", func(t *testing.T) {
		t.Cleanup(restore)
		e.Validator = &stubValidator{err: errors.New("v")}
		ctx, rec := newUpdateCtx(e, "1", "name=a&email=a@b.com")
		err := UpdateUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "v")
	})

	t.Run("bad email", func(t *testing.T) {
		t.Cleanup(restore)
		e.Validator = &stubValidator{}
		ctx, rec := newUpdateCtx(e, "1", "name=a&email=bad")
		err := UpdateUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("update error", func(t *testing.T) {
		t.Cleanup(restore)
		e.Validator = &stubValidator{}
		updateUser = func(context.Context, database.DB, *model.User) error { return errors.New("u") }
		ctx, rec := newUpdateCtx(e, "1", "name=a&email=a@b.com")
		err := UpdateUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Cleanup(restore)
		e.Validator = &stubValidator{}
		var got model.User
		updateUser = func(_ context.Context, _ database.DB, u *model.User) error {
			got = *u
			return nil
		}
		ctx, rec := newUpdateCtx(e, "2", "name=A&email=B@EX.com")
		err := UpdateUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, rec.Code)
		require.Equal(t, "b@ex.com", got.Email)
		require.Equal(t, 2, got.ID)
	})
}

func TestDeleteUserHandler(t *testing.T) {
	e := echo.New()
	t.Run("bad id", func(t *testing.T) {
		t.Cleanup(restore)
		ctx, rec := newParamCtx(e, "x")
		err := DeleteUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("delete error", func(t *testing.T) {
		t.Cleanup(restore)
		deleteUser = func(context.Context, database.DB, int) error { return errors.New("d") }
		ctx, rec := newParamCtx(e, "1")
		err := DeleteUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Cleanup(restore)
		deleteUser = func(context.Context, database.DB, int) error { return nil }
		ctx, rec := newParamCtx(e, "2")
		err := DeleteUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, rec.Code)
	})
}

func TestGetMyUserHandler(t *testing.T) {
	e := echo.New()
	t.Run("no claims", func(t *testing.T) {
		t.Cleanup(restore)
		ctx, rec := newMeCtx(e, http.MethodGet, "")
		err := GetMyUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("get error", func(t *testing.T) {
		t.Cleanup(restore)
		getUserByID = func(context.Context, database.DB, int) (*model.User, error) { return nil, errors.New("e") }
		ctx, rec := newMeCtx(e, http.MethodGet, "")
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := GetMyUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Cleanup(restore)
		now := time.Now().UTC()
		getUserByID = func(context.Context, database.DB, int) (*model.User, error) {
			return &model.User{ID: 1, Name: "n", Email: "e", CreatedAt: now}, nil
		}
		ctx, rec := newMeCtx(e, http.MethodGet, "")
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := GetMyUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), "\"id\":1")
	})
}

func TestUpdateMyUserHandler(t *testing.T) {
	e := echo.New()
	e.Validator = &stubValidator{}

	t.Run("bind error", func(t *testing.T) {
		t.Cleanup(restore)
		ctx, rec := newMeCtx(e, http.MethodPut, "%")
		err := UpdateMyUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("validate error", func(t *testing.T) {
		t.Cleanup(restore)
		e.Validator = &stubValidator{err: errors.New("v")}
		ctx, rec := newMeCtx(e, http.MethodPut, "name=a&email=a@b.com")
		err := UpdateMyUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "v")
	})

	t.Run("no claims", func(t *testing.T) {
		t.Cleanup(restore)
		e.Validator = &stubValidator{}
		ctx, rec := newMeCtx(e, http.MethodPut, "name=a&email=a@b.com")
		err := UpdateMyUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("bad email", func(t *testing.T) {
		t.Cleanup(restore)
		ctx, rec := newMeCtx(e, http.MethodPut, "name=a&email=bad")
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := UpdateMyUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("update error", func(t *testing.T) {
		t.Cleanup(restore)
		updateUser = func(context.Context, database.DB, *model.User) error { return errors.New("u") }
		ctx, rec := newMeCtx(e, http.MethodPut, "name=a&email=a@b.com")
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := UpdateMyUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Cleanup(restore)
		var got model.User
		updateUser = func(_ context.Context, _ database.DB, u *model.User) error {
			got = *u
			return nil
		}
		ctx, rec := newMeCtx(e, http.MethodPut, "name=A&email=B@Ex.com")
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 5})
		err := UpdateMyUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, rec.Code)
		require.Equal(t, 5, got.ID)
		require.Equal(t, "b@ex.com", got.Email)
	})
}

func TestUpdateMyUserPasswordHandler(t *testing.T) {
	e := echo.New()
	e.Validator = &stubValidator{}

	form := "old_password=o&new_password=n"

	t.Run("bind error", func(t *testing.T) {
		t.Cleanup(restore)
		ctx, rec := newMeCtx(e, http.MethodPatch, "%")
		err := UpdateMyUserPasswordHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("validate error", func(t *testing.T) {
		t.Cleanup(restore)
		e.Validator = &stubValidator{err: errors.New("v")}
		ctx, rec := newMeCtx(e, http.MethodPatch, form)
		err := UpdateMyUserPasswordHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("no claims", func(t *testing.T) {
		t.Cleanup(restore)
		e.Validator = &stubValidator{}
		ctx, rec := newMeCtx(e, http.MethodPatch, form)
		err := UpdateMyUserPasswordHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("get error", func(t *testing.T) {
		t.Cleanup(restore)
		getUserByID = func(context.Context, database.DB, int) (*model.User, error) { return nil, errors.New("g") }
		ctx, rec := newMeCtx(e, http.MethodPatch, form)
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := UpdateMyUserPasswordHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("auth fail", func(t *testing.T) {
		t.Cleanup(restore)
		getUserByID = func(context.Context, database.DB, int) (*model.User, error) { return &model.User{ID: 1}, nil }
		authenticateUser = func(context.Context, model.User, string) error { return errors.New("bad") }
		ctx, rec := newMeCtx(e, http.MethodPatch, form)
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := UpdateMyUserPasswordHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("hash error", func(t *testing.T) {
		t.Cleanup(restore)
		getUserByID = func(context.Context, database.DB, int) (*model.User, error) { return &model.User{ID: 1}, nil }
		authenticateUser = func(context.Context, model.User, string) error { return nil }
		hashPassword = func(string) (string, error) { return "", errors.New("h") }
		ctx, rec := newMeCtx(e, http.MethodPatch, form)
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := UpdateMyUserPasswordHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("update error", func(t *testing.T) {
		t.Cleanup(restore)
		getUserByID = func(context.Context, database.DB, int) (*model.User, error) { return &model.User{ID: 1}, nil }
		authenticateUser = func(context.Context, model.User, string) error { return nil }
		hashPassword = func(string) (string, error) { return "h", nil }
		updateUserPassword = func(context.Context, database.DB, int, string) error { return errors.New("u") }
		ctx, rec := newMeCtx(e, http.MethodPatch, form)
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := UpdateMyUserPasswordHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Cleanup(restore)
		var updatedID int
		getUserByID = func(context.Context, database.DB, int) (*model.User, error) { return &model.User{ID: 1}, nil }
		authenticateUser = func(context.Context, model.User, string) error { return nil }
		hashPassword = func(string) (string, error) { return "h", nil }
		updateUserPassword = func(_ context.Context, _ database.DB, id int, _ string) error {
			updatedID = id
			return nil
		}
		ctx, rec := newMeCtx(e, http.MethodPatch, form)
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 9})
		err := UpdateMyUserPasswordHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, rec.Code)
		require.Equal(t, 9, updatedID)
	})
}

func TestDeleteMyUserHandler(t *testing.T) {
	e := echo.New()
	t.Run("no claims", func(t *testing.T) {
		t.Cleanup(restore)
		ctx, rec := newMeCtx(e, http.MethodDelete, "")
		err := DeleteMyUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("delete error", func(t *testing.T) {
		t.Cleanup(restore)
		deleteUser = func(context.Context, database.DB, int) error { return errors.New("d") }
		ctx, rec := newMeCtx(e, http.MethodDelete, "")
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 1})
		err := DeleteMyUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Cleanup(restore)
		deleteUser = func(context.Context, database.DB, int) error { return nil }
		ctx, rec := newMeCtx(e, http.MethodDelete, "")
		ctx.Set(middleware.ContextUserKey, &service.CustomClaims{UserID: 2})
		err := DeleteMyUserHandler(nil)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, rec.Code)
	})
}
