package db

import (
	"testing"

	"github.com/psto/irw/internal/config"
)

func TestCreateTables(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := Connect(config.NullConfig{}, tmpDir+"/test.db")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer db.Close()

	if err := CreateTables(db); err != nil {
		t.Fatalf("CreateTables failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='tracks'").Scan(&count); err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("tracks table not created, got count %d", count)
	}

	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='sessions'").Scan(&count); err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("sessions table not created, got count %d", count)
	}
}
