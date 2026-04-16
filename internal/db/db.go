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
	_, err := Connect(dbPath)
	return err
}
