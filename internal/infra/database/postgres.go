package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// ensureCompatibleWithPooler ajusta a URL para pgx + PgBouncer (ex.: Supabase transaction pooler :6543).
// Sem isso, o cache de prepared statements do pgx pode falhar com SQLSTATE 42P05 ("prepared statement ... already exists").
func ensureCompatibleWithPooler(databaseURL string) (string, error) {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return "", fmt.Errorf("parse database URL: %w", err)
	}
	q := u.Query()
	if q.Get("default_query_exec_mode") == "" {
		// pgx v5 aceita: cache_statement | cache_describe | describe_exec | exec | simple_protocol
		// simple_protocol evita prepared statements no wire — compatível com PgBouncer modo transação (ex.: Supabase :6543).
		q.Set("default_query_exec_mode", "simple_protocol")
		u.RawQuery = q.Encode()
	}
	return u.String(), nil
}

func Open(databaseURL string) (*sql.DB, error) {
	fixedURL, err := ensureCompatibleWithPooler(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("database.Open: %w", err)
	}

	db, err := sql.Open("pgx", fixedURL)
	if err != nil {
		return nil, fmt.Errorf("database.Open: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database.Open: ping: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	slog.Info("database connected", "driver", "postgres")
	return db, nil
}

func RunMigrations(db *sql.DB, migrationsPath string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("database.RunMigrations: driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("database.RunMigrations: init: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("database.RunMigrations: up: %w", err)
	}

	version, dirty, _ := m.Version()
	slog.Info("migrations applied", "version", version, "dirty", dirty)
	return nil
}
