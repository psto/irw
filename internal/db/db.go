package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/psto/irw/internal/config"

	_ "modernc.org/sqlite"
)

func Connect(cfg config.ConfigProvider, dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		dbPath = cfg.GetDBPath()
	}

	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func CreateTables(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS tracks (

		path TEXT PRIMARY KEY,
		type TEXT DEFAULT 'reading',
		interval REAL DEFAULT 1.0,
		afactor REAL DEFAULT 2.0,
		due_date INTEGER DEFAULT (strftime('%s', 'now')),
		is_finished INTEGER DEFAULT 0,
		priority REAL DEFAULT (ABS(random() % 21) + 40)

	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
	 CREATE TABLE IF NOT EXISTS sessions (

		date TEXT DEFAULT (date('now', 'localtime')),
		duration INTEGER,
		reviewed INTEGER,
		finished INTEGER

	 )`)
	return err
}
