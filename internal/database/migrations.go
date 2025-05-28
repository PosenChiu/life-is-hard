// File: internal/database/migrations.go
package database

import (
	"database/sql"
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations 嵌入並執行 SQL migration (up all)
func RunMigrations(dbURL string) error {
	// 建立 *sql.DB 使用 pgx stdlib driver
	sqlDB, err := sql.Open("pgx", dbURL)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	// 建立 migrate driver for postgres
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return err
	}

	// embed migrations from migrationsFS
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	// 初始化 migrate
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return err
	}

	// 執行升級到最新版本
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// RollbackAll 退回所有 migration (down to version 0)
func RollbackAll(dbURL string) error {
	// 建立 *sql.DB 使用 pgx stdlib driver
	sqlDB, err := sql.Open("pgx", dbURL)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	// 建立 migrate driver for postgres
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return err
	}

	// embed migrations from migrationsFS
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	// 初始化 migrate
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return err
	}

	// 執行回滾到底
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
