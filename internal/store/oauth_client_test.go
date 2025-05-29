package store

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

// fakeRow 實作 pgx.Row，用於模擬單筆掃描行為。
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
		// GetOAuthClientByClientID: client_id, client_secret, user_id, grant_types, created_at, updated_at
		*dest[0].(*string) = c.ClientID
		*dest[1].(*string) = c.ClientSecret
		*dest[2].(*int) = c.UserID
		*dest[3].(*[]string) = c.GrantTypes
		*dest[4].(*time.Time) = c.CreatedAt
		*dest[5].(*time.Time) = c.UpdatedAt
	case 3:
		// CreateOAuthClient: client_id, created_at, updated_at
		*dest[0].(*string) = c.ClientID
		*dest[1].(*time.Time) = c.CreatedAt
		*dest[2].(*time.Time) = c.UpdatedAt
	case 1:
		// UpdateOAuthClient: updated_at
		*dest[0].(*time.Time) = c.UpdatedAt
	default:
		panic("fakeRow.Scan: unexpected number of dest")
	}
	return nil
}

// fakeRows 實作 pgx.Rows，用於模擬多筆掃描行為。
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
func (r *fakeRows) Next() bool                                   { return r.idx < len(r.data) }
func (r *fakeRows) Scan(dest ...any) error {
	if r.scanErr != nil {
		return r.scanErr
	}
	c := r.data[r.idx]
	r.idx++
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

/* ---------- 完整測試 ---------- */

func TestOAuthClientRepository(t *testing.T) {
	now := time.Now().UTC()
	sample := model.OAuthClient{
		ClientID:     "cid",
		ClientSecret: "sec",
		UserID:       1,
		GrantTypes:   []string{"password"},
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	/* GetOAuthClientByClientID */
	t.Run("Get ok", func(t *testing.T) {
		p := &database.FakeDB{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return &fakeRow{client: &sample}
			},
		}
		got, err := GetOAuthClientByClientID(context.Background(), p, "cid")
		require.NoError(t, err)
		require.Equal(t, sample.ClientID, got.ClientID)
	})

	t.Run("Get err", func(t *testing.T) {
		p := &database.FakeDB{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return &fakeRow{scanErr: errors.New("not found")}
			},
		}
		_, err := GetOAuthClientByClientID(context.Background(), p, "cid")
		require.Error(t, err)
	})

	/* CreateOAuthClient */
	t.Run("Create ok", func(t *testing.T) {
		p := &database.FakeDB{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return &fakeRow{client: &sample}
			},
		}
		require.NoError(t, CreateOAuthClient(context.Background(), p, &sample))
	})

	t.Run("Create err", func(t *testing.T) {
		p := &database.FakeDB{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return &fakeRow{scanErr: errors.New("dup")}
			},
		}
		require.Error(t, CreateOAuthClient(context.Background(), p, &sample))
	})

	/* UpdateOAuthClient */
	t.Run("Update ok", func(t *testing.T) {
		p := &database.FakeDB{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
			// QueryRow also used for Update returning
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return &fakeRow{client: &sample}
			},
		}
		require.NoError(t, UpdateOAuthClient(context.Background(), p, &sample))
	})

	t.Run("Update err", func(t *testing.T) {
		p := &database.FakeDB{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return &fakeRow{scanErr: errors.New("fail update")}
			},
		}
		require.Error(t, UpdateOAuthClient(context.Background(), p, &sample))
	})

	/* DeleteOAuthClient */
	t.Run("Delete ok", func(t *testing.T) {
		p := &database.FakeDB{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
		}
		require.NoError(t, DeleteOAuthClient(context.Background(), p, "cid"))
	})

	t.Run("Delete err", func(t *testing.T) {
		p := &database.FakeDB{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, errors.New("fail delete")
			},
		}
		require.Error(t, DeleteOAuthClient(context.Background(), p, "cid"))
	})

	/* ListOAuthClients */
	t.Run("List ok", func(t *testing.T) {
		rows := &fakeRows{data: []model.OAuthClient{sample, sample}}
		p := &database.FakeDB{
			QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
				return rows, nil
			},
		}
		list, err := ListOAuthClients(context.Background(), p, 1)
		require.NoError(t, err)
		require.Len(t, list, 2)
	})

	t.Run("List query err", func(t *testing.T) {
		p := &database.FakeDB{
			QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
				return nil, errors.New("database fail")
			},
		}
		_, err := ListOAuthClients(context.Background(), p, 1)
		require.Error(t, err)
	})

	t.Run("List scan err", func(t *testing.T) {
		rows := &fakeRows{data: []model.OAuthClient{sample}, scanErr: errors.New("scan fail")}
		p := &database.FakeDB{
			QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
				return rows, nil
			},
		}
		_, err := ListOAuthClients(context.Background(), p, 1)
		require.Error(t, err)
	})
}
