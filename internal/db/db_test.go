package db

import (
	"os"
	"testing"
)

func TestConnect(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	db, err := Connect(dbPath)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer db.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file was not created")
	}
}
