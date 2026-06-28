package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// Open opens the SQLite database at the given path and enables WAL mode.
func Open(path string) error {
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	// SQLite performs best with a single writer connection.
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}
	DB = db
	return nil
}

// Close closes the database connection.
func Close() {
	if DB != nil {
		_ = DB.Close()
		DB = nil
	}
}
