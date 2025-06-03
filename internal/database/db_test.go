package database

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

type fakeRows struct{}

func (fakeRows) Close()                                       {}
func (fakeRows) Err() error                                   { return nil }
func (fakeRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (fakeRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (fakeRows) Next() bool                                   { return false }
func (fakeRows) Scan(dest ...any) error                       { return nil }
func (fakeRows) Values() ([]any, error)                       { return nil, nil }
func (fakeRows) RawValues() [][]byte                          { return nil }
func (fakeRows) Conn() *pgx.Conn                              { return nil }

func TestFakeDB(t *testing.T) {
	db := &FakeDB{}
	require.Panics(t, func() { db.Exec(context.Background(), "", nil) })
	require.Panics(t, func() { db.Query(context.Background(), "") })
	require.Panics(t, func() { db.QueryRow(context.Background(), "") })
	require.Panics(t, func() { db.Ping(context.Background()) })
	db.Close()

	execCalled := false
	queryCalled := false
	rowCalled := false
	pingCalled := false
	closeCalled := false

	db.ExecFn = func(ctx context.Context, s string, args ...any) (pgconn.CommandTag, error) {
		execCalled = true
		return pgconn.CommandTag{}, errors.New("e")
	}
	db.QueryFn = func(ctx context.Context, s string, args ...any) (pgx.Rows, error) {
		queryCalled = true
		return fakeRows{}, nil
	}
	db.QueryRowFn = func(ctx context.Context, s string, args ...any) pgx.Row {
		rowCalled = true
		return pgx.Row(fakeRows{})
	}
	db.PingFn = func(ctx context.Context) error { pingCalled = true; return nil }
	db.CloseFn = func() { closeCalled = true }

	_, err := db.Exec(context.Background(), "sql")
	require.Error(t, err)
	_, err = db.Query(context.Background(), "sql")
	require.NoError(t, err)
	_ = db.QueryRow(context.Background(), "sql")
	require.NoError(t, db.Ping(context.Background()))
	db.Close()
	require.True(t, execCalled)
	require.True(t, queryCalled)
	require.True(t, rowCalled)
	require.True(t, pingCalled)
	require.True(t, closeCalled)
}
