// Package sqlite handles the connection to the SQLite database and the
// application of database migrations. This is the "Database" part of the
// project (connection + migrations), kept separate from the App logic.
package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	migratesqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

// Open opens (creating if needed) the SQLite database at the given path and
// returns a ready-to-use *sql.DB with foreign keys enabled.
func Open(dbPath string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	// _foreign_keys=on enforces FK constraints; busy_timeout avoids "database
	// is locked" errors when websocket and HTTP handlers write concurrently.
	dsn := fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL", dbPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	// SQLite handles one writer at a time; a single connection avoids lock churn.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}

// RunMigrations applies every "up" migration found in the migrations folder.
// It is idempotent: already-applied migrations are skipped on subsequent runs.
func RunMigrations(db *sql.DB, migrationsPath string) error {
	resolved, err := resolveMigrationsPath(migrationsPath)
	if err != nil {
		return err
	}

	driver, err := migratesqlite3.WithInstance(db, &migratesqlite3.Config{})
	if err != nil {
		return fmt.Errorf("migrate driver: %w", err)
	}

	sourceURL := "file://" + filepath.ToSlash(resolved)
	m, err := migrate.NewWithDatabaseInstance(sourceURL, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("migrate init (%s): %w", sourceURL, err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}

	version, _, _ := m.Version()
	log.Printf("migrations applied, schema version: %d", version)
	return nil
}

// resolveMigrationsPath returns an absolute path to the migrations directory,
// trying the configured path first and then a few common fallbacks so the
// server works whether it is started from the repo root or the backend folder.
func resolveMigrationsPath(p string) (string, error) {
	candidates := []string{
		p,
		filepath.Join("backend", p),
		filepath.Join("..", p),
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			abs, err := filepath.Abs(c)
			if err != nil {
				return "", err
			}
			return abs, nil
		}
	}
	return "", fmt.Errorf("migrations directory not found (looked for %q)", p)
}
