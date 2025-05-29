package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Ping(context.Context) error
	Close()
}

type FakeDB struct {
	ExecFn     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryFn    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRowFn func(ctx context.Context, sql string, args ...any) pgx.Row
	PingFn     func(ctx context.Context) error
	CloseFn    func()
}

func (f *FakeDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if f.ExecFn != nil {
		return f.ExecFn(ctx, sql, args...)
	}
	panic("unexpected Exec")
}

func (f *FakeDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if f.QueryFn != nil {
		return f.QueryFn(ctx, sql, args...)
	}
	panic("unexpected Query")
}

func (f *FakeDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if f.QueryRowFn != nil {
		return f.QueryRowFn(ctx, sql, args...)
	}
	panic("unexpected QueryRow")
}

func (f *FakeDB) Ping(ctx context.Context) error {
	if f.QueryRowFn != nil {
		return f.PingFn(ctx)
	}
	panic("unexpected QueryRow")
}

func (f *FakeDB) Close() {
	if f.CloseFn != nil {
		f.CloseFn()
	}
}
