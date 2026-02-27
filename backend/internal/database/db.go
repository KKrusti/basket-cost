// Package database manages the SQLite connection and schema migrations.
package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Open opens (or creates) the SQLite database at the given path,
// applies performance PRAGMAs, and runs all schema migrations.
// Pass ":memory:" for an in-memory database (useful in tests).
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// SQLite is not safe for concurrent writes from multiple connections;
	// a single connection avoids locking issues without a connection pool.
	db.SetMaxOpenConns(1)

	if err := applyPragmas(db); err != nil {
		db.Close()
		return nil, err
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// applyPragmas configures performance and integrity settings.
func applyPragmas(db *sql.DB) error {
	pragmas := `
		PRAGMA foreign_keys  = ON;
		PRAGMA journal_mode  = WAL;
		PRAGMA synchronous   = NORMAL;
		PRAGMA temp_store    = MEMORY;
		PRAGMA cache_size    = -16000;
	`
	if _, err := db.Exec(pragmas); err != nil {
		return fmt.Errorf("apply pragmas: %w", err)
	}
	return nil
}

// migrate creates all required tables if they do not exist.
// Add new migrations as numbered steps below; never alter existing ones.
func migrate(db *sql.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS products (
			id       TEXT PRIMARY KEY,
			name     TEXT NOT NULL,
			category TEXT NOT NULL DEFAULT ''
		);

		CREATE TABLE IF NOT EXISTS price_records (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			product_id TEXT    NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			date       TEXT    NOT NULL,  -- ISO-8601: YYYY-MM-DD
			price      REAL    NOT NULL,
			store      TEXT    NOT NULL DEFAULT ''
		);

		CREATE INDEX IF NOT EXISTS idx_price_records_product_id
			ON price_records(product_id);
	`
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}
	return nil
}
