// File: internal/database/postgres.go
package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func NewPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	return pgxpool.NewWithConfig(ctx, cfg)
}

type FakePool struct {
	QueryRowFn func(ctx context.Context, sql string, args ...any) pgx.Row
	QueryFn    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	ExecFn     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (p *FakePool) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if p.QueryRowFn != nil {
		return p.QueryRowFn(ctx, query, args...)
	}
	panic("unexpected QueryRow")
}

func (p *FakePool) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	if p.QueryFn != nil {
		return p.QueryFn(ctx, query, args...)
	}
	panic("unexpected Query")
}

func (p *FakePool) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	if p.ExecFn != nil {
		return p.ExecFn(ctx, query, args...)
	}
	panic("unexpected Exec")
}
