package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var defaultDBPath string

func init() {
	home, _ := os.UserHomeDir()
	defaultDBPath = filepath.Join(home, ".local/share/ir-tool/ir.db")
}

func Connect(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		dbPath = defaultDBPath
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

func Init(dbPath string) error {
	db, err := Connect(dbPath)
	if err != nil {
		return err
	}
	return CreateTables(db)
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
