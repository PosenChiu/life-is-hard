package database

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	dbdriver "github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	src "github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type fakeMigrator struct{ upErr, downErr error }

func (f fakeMigrator) Up() error   { return f.upErr }
func (f fakeMigrator) Down() error { return f.downErr }

func restore() {
	pgxpoolNew = pgxpool.New
	sqlOpenDB = sql.Open
	postgresWithInstanceFn = postgres.WithInstance
	iofsNewFn = iofs.New
	migrateNewWithInstance = func(sourceName string, sourceDriver src.Driver, databaseName string, databaseDriver dbdriver.Driver) (migrateInstance, error) {
		m, err := migrate.NewWithInstance(sourceName, sourceDriver, databaseName, databaseDriver)
		if err != nil {
			return nil, err
		}
		return m, nil
	}
}

func TestNewPgxPool(t *testing.T) {
	t.Cleanup(restore)
	pgxpoolNew = func(ctx context.Context, url string) (*pgxpool.Pool, error) { return nil, errors.New("bad") }
	_, err := NewPgxPool(context.Background(), "url")
	require.Error(t, err)

	pgxpoolNew = func(ctx context.Context, url string) (*pgxpool.Pool, error) { return &pgxpool.Pool{}, nil }
	db, err := NewPgxPool(context.Background(), "url")
	require.NoError(t, err)
	require.NotNil(t, db)
}

func TestRunMigrationsAndRollback(t *testing.T) {
	t.Cleanup(restore)
	sqlOpenDB = func(driver, dsn string) (*sql.DB, error) { return nil, errors.New("open") }
	require.Error(t, RunMigrations("url"))
	sqlOpenDB = func(driver, dsn string) (*sql.DB, error) { return sql.Open("pgx", "") }
	postgresWithInstanceFn = func(*sql.DB, *postgres.Config) (dbdriver.Driver, error) { return nil, errors.New("drv") }
	require.Error(t, RunMigrations("url"))

	postgresWithInstanceFn = func(*sql.DB, *postgres.Config) (dbdriver.Driver, error) { return nil, nil }
	iofsNewFn = func(f fs.FS, s string) (src.Driver, error) { return nil, errors.New("src") }
	require.Error(t, RunMigrations("url"))

	iofsNewFn = func(f fs.FS, s string) (src.Driver, error) { return nil, nil }
	migrateNewWithInstance = func(string, src.Driver, string, dbdriver.Driver) (migrateInstance, error) {
		return nil, errors.New("mig")
	}
	require.Error(t, RunMigrations("url"))

	migrateNewWithInstance = func(string, src.Driver, string, dbdriver.Driver) (migrateInstance, error) {
		return fakeMigrator{upErr: errors.New("u")}, nil
	}
	require.Error(t, RunMigrations("url"))

	migrateNewWithInstance = func(string, src.Driver, string, dbdriver.Driver) (migrateInstance, error) {
		return fakeMigrator{upErr: migrate.ErrNoChange}, nil
	}
	require.NoError(t, RunMigrations("url"))

	migrateNewWithInstance = func(string, src.Driver, string, dbdriver.Driver) (migrateInstance, error) { return fakeMigrator{}, nil }
	require.NoError(t, RollbackAll("url"))

	sqlOpenDB = func(string, string) (*sql.DB, error) { return nil, errors.New("open") }
	require.Error(t, RollbackAll("url"))
	sqlOpenDB = func(string, string) (*sql.DB, error) { return sql.Open("pgx", "") }
	postgresWithInstanceFn = func(*sql.DB, *postgres.Config) (dbdriver.Driver, error) { return nil, errors.New("drv") }
	require.Error(t, RollbackAll("url"))
	postgresWithInstanceFn = func(*sql.DB, *postgres.Config) (dbdriver.Driver, error) { return nil, nil }
	iofsNewFn = func(fs.FS, string) (src.Driver, error) { return nil, errors.New("src") }
	require.Error(t, RollbackAll("url"))
	iofsNewFn = func(fs.FS, string) (src.Driver, error) { return nil, nil }
	migrateNewWithInstance = func(string, src.Driver, string, dbdriver.Driver) (migrateInstance, error) {
		return nil, errors.New("mig")
	}
	require.Error(t, RollbackAll("url"))
	migrateNewWithInstance = func(string, src.Driver, string, dbdriver.Driver) (migrateInstance, error) {
		return fakeMigrator{downErr: migrate.ErrNoChange}, nil
	}
	require.NoError(t, RollbackAll("url"))
	migrateNewWithInstance = func(string, src.Driver, string, dbdriver.Driver) (migrateInstance, error) {
		return fakeMigrator{downErr: errors.New("d")}, nil
	}
	require.Error(t, RollbackAll("url"))
}
