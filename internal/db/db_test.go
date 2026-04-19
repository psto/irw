package db

import (
	"os"
	"testing"

	"github.com/psto/irw/internal/config"
)

func TestConnect(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	db, err := Connect(config.NullConfig{}, dbPath)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer db.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file was not created")
	}
}

func TestConnectWithFlagPath(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"
	cfg := config.NullConfig{}

	db, err := Connect(cfg, dbPath)
	if err != nil {
		t.Fatalf("Connect with flag path failed: %v", err)
	}
	db.Close()
}

func TestConnectCreatesDir(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/subdir/test.db"
	cfg := config.NullConfig{}

	db, err := Connect(cfg, dbPath)
	if err != nil {
		t.Fatalf("Connect failed to create dir: %v", err)
	}
	db.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file was not created")
	}
}

func TestConnectWithRealConfig(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	origConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origConfigHome)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	origDataHome := os.Getenv("XDG_DATA_HOME")
	defer os.Setenv("XDG_DATA_HOME", origDataHome)
	os.Setenv("XDG_DATA_HOME", tmpDir)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load config failed: %v", err)
	}

	db, err := Connect(cfg, dbPath)
	if err != nil {
		t.Fatalf("Connect with real config failed: %v", err)
	}
	db.Close()
}
