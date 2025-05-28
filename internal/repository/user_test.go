// File: internal/repository/user_test.go
package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"life-is-hard/internal/database"
	"life-is-hard/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

/* ---------- 假實作 ---------- */

// fakeUserRow 支援兩種 Scan 呼叫場景：
// 1) len(dest)==6 → GetUserByID / GetUserByName
// 2) len(dest)==2 → CreateUser (id, created_at)
type fakeUserRow struct {
	scanErr error
	user    *model.User
}

func (r *fakeUserRow) Scan(dest ...any) error {
	if r.scanErr != nil {
		return r.scanErr
	}
	u := r.user
	switch len(dest) {
	case 6:
		*dest[0].(*int) = u.ID
		*dest[1].(*string) = u.Name
		*dest[2].(*string) = u.Email
		*dest[3].(*string) = u.PasswordHash
		*dest[4].(*time.Time) = u.CreatedAt
		*dest[5].(*bool) = u.IsAdmin
	case 2:
		*dest[0].(*int) = u.ID
		*dest[1].(*time.Time) = u.CreatedAt
	default:
		panic("fakeUserRow.Scan: unexpected dest count")
	}
	return nil
}

/* ---------- 完整測試 ---------- */

func TestUserRepository(t *testing.T) {
	now := time.Now().UTC()
	sample := &model.User{
		ID:           7,
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "hash123",
		CreatedAt:    now,
		IsAdmin:      true,
	}

	/* --- GetUserByID --- */
	t.Run("GetUserByID success", func(t *testing.T) {
		p := &database.FakePool{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return &fakeUserRow{user: sample}
			},
		}
		u, err := GetUserByID(context.Background(), p, 7)
		require.NoError(t, err)
		require.Equal(t, sample.Email, u.Email)
		require.True(t, u.IsAdmin)
	})

	t.Run("GetUserByID not found", func(t *testing.T) {
		p := &database.FakePool{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return &fakeUserRow{scanErr: errors.New("no rows")}
			},
		}
		u, err := GetUserByID(context.Background(), p, 999)
		require.Error(t, err)
		require.Nil(t, u)
	})

	/* --- GetUserByName --- */
	t.Run("GetUserByName success", func(t *testing.T) {
		p := &database.FakePool{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return &fakeUserRow{user: sample}
			},
		}
		u, err := GetUserByName(context.Background(), p, "Alice")
		require.NoError(t, err)
		require.Equal(t, 7, u.ID)
	})

	t.Run("GetUserByName not found", func(t *testing.T) {
		p := &database.FakePool{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return &fakeUserRow{scanErr: errors.New("no rows")}
			},
		}
		u, err := GetUserByName(context.Background(), p, "Bob")
		require.Error(t, err)
		require.Nil(t, u)
	})

	/* --- CreateUser --- */
	t.Run("CreateUser success", func(t *testing.T) {
		newUser := &model.User{Name: "Bob", Email: "bob@example.com", PasswordHash: "pwdhash", IsAdmin: false}
		p := &database.FakePool{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				u := *newUser
				u.ID = 42
				u.CreatedAt = now.Add(time.Hour)
				return &fakeUserRow{user: &u}
			},
		}
		created, err := CreateUser(context.Background(), p, newUser)
		require.NoError(t, err)
		require.Equal(t, 42, created.ID)
		require.WithinDuration(t, now.Add(time.Hour), created.CreatedAt, time.Second)
	})

	t.Run("CreateUser error", func(t *testing.T) {
		p := &database.FakePool{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return &fakeUserRow{scanErr: errors.New("dup key")}
			},
		}
		_, err := CreateUser(context.Background(), p, &model.User{})
		require.Error(t, err)
	})

	/* --- UpdateUser --- */
	t.Run("UpdateUser success", func(t *testing.T) {
		p := &database.FakePool{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
		}
		err := UpdateUser(context.Background(), p, sample)
		require.NoError(t, err)
	})

	t.Run("UpdateUser error", func(t *testing.T) {
		p := &database.FakePool{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, errors.New("update failed")
			},
		}
		err := UpdateUser(context.Background(), p, sample)
		require.Error(t, err)
	})

	/* --- UpdateUserPassword --- */
	t.Run("UpdateUserPassword success", func(t *testing.T) {
		p := &database.FakePool{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
		}
		err := UpdateUserPassword(context.Background(), p, 7, "newHash")
		require.NoError(t, err)
	})

	t.Run("UpdateUserPassword error", func(t *testing.T) {
		p := &database.FakePool{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, errors.New("pwd update failed")
			},
		}
		err := UpdateUserPassword(context.Background(), p, 7, "newHash")
		require.Error(t, err)
	})

	/* --- DeleteUser --- */
	t.Run("DeleteUser success", func(t *testing.T) {
		p := &database.FakePool{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
		}
		err := DeleteUser(context.Background(), p, 7)
		require.NoError(t, err)
	})

	t.Run("DeleteUser error", func(t *testing.T) {
		p := &database.FakePool{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, errors.New("delete failed")
			},
		}
		err := DeleteUser(context.Background(), p, 7)
		require.Error(t, err)
	})
}
