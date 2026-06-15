package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// Open opens (or creates) a SQLite database at path, creating the parent
// directory if missing and running the embedded schema. The schema uses
// CREATE TABLE IF NOT EXISTS so re-opens are a no-op.
func Open(path string) (*sql.DB, error) {
	if path == "" {
		return nil, fmt.Errorf("db path is empty")
	}
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db dir %q: %w", dir, err)
		}
	}
	dsn := path + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %q: %w", path, err)
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("run schema: %w", err)
	}
	// Migrations for columns added after a table first shipped. schema.sql's
	// CREATE TABLE IF NOT EXISTS is a no-op on existing DBs, so new columns must
	// be added explicitly. ensureColumn is idempotent.
	if err := ensureColumn(db, "users", "password_hash", "TEXT"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate users.password_hash: %w", err)
	}
	return db, nil
}

// ensureColumn adds `column ddl` to `table` if it isn't already present. Safe to
// run on every boot. (table/column/ddl are internal constants, never user input.)
func ensureColumn(db *sql.DB, table, column, ddl string) error {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			cid, notnull, pk int
			name, ctype      string
			dflt             sql.NullString
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if name == column {
			return nil // already present
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	_, err = db.Exec("ALTER TABLE " + table + " ADD COLUMN " + column + " " + ddl)
	return err
}
