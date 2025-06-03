package database

import (
	"context"
	"database/sql"
	"embed"

	"github.com/golang-migrate/migrate/v4"
	dbdriver "github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	src "github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	pgxpoolNew             = pgxpool.New
	sqlOpenDB              = sql.Open
	postgresWithInstanceFn = postgres.WithInstance
	iofsNewFn              = iofs.New
	migrateNewWithInstance = func(sourceName string, sourceDriver src.Driver, databaseName string, databaseDriver dbdriver.Driver) (migrateInstance, error) {
		m, err := migrate.NewWithInstance(sourceName, sourceDriver, databaseName, databaseDriver)
		if err != nil {
			return nil, err
		}
		return m, nil
	}
)

type migrateInstance interface {
	Up() error
	Down() error
}

func NewPgxPool(ctx context.Context, url string) (DB, error) {
	pool, err := pgxpoolNew(ctx, url)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations 嵌入並執行 SQL migration (up all)
func RunMigrations(dbURL string) error {
	sqlDB, err := sqlOpenDB("pgx", dbURL)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	driver, err := postgresWithInstanceFn(sqlDB, &postgres.Config{})
	if err != nil {
		return err
	}

	sourceDriver, err := iofsNewFn(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	m, err := migrateNewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// RollbackAll 退回所有 migration (down to version 0)
func RollbackAll(dbURL string) error {
	sqlDB, err := sqlOpenDB("pgx", dbURL)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	driver, err := postgresWithInstanceFn(sqlDB, &postgres.Config{})
	if err != nil {
		return err
	}

	sourceDriver, err := iofsNewFn(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	m, err := migrateNewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
