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
	// Migration 1: base schema.
	m1 := `
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
	if _, err := db.Exec(m1); err != nil {
		return fmt.Errorf("migrate m1: %w", err)
	}

	// Migration 2: add image_url column (idempotent via ALTER TABLE IF NOT EXISTS column pattern).
	// SQLite does not support IF NOT EXISTS for ADD COLUMN, so we check the column first.
	var colCount int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM pragma_table_info('products') WHERE name='image_url'`,
	).Scan(&colCount)
	if err != nil {
		return fmt.Errorf("migrate m2 check: %w", err)
	}
	if colCount == 0 {
		if _, err := db.Exec(`ALTER TABLE products ADD COLUMN image_url TEXT NOT NULL DEFAULT ''`); err != nil {
			return fmt.Errorf("migrate m2 add column: %w", err)
		}
	}

	// Migration 3: track processed PDF filenames to prevent duplicate imports.
	m3 := `
		CREATE TABLE IF NOT EXISTS processed_files (
			filename    TEXT PRIMARY KEY,
			imported_at TEXT NOT NULL   -- ISO-8601 timestamp
		);
	`
	if _, err := db.Exec(m3); err != nil {
		return fmt.Errorf("migrate m3: %w", err)
	}

	return nil
}
