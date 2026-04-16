package db

import (
	"testing"
)

func TestLogSession(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := Connect(tmpDir + "/test.db")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer db.Close()
	CreateTables(db)

	err = LogSession(db, 300, 5, 2)
	if err != nil {
		t.Fatalf("LogSession failed: %v", err)
	}
}

func TestGetRecentSessions(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := Connect(tmpDir + "/test.db")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer db.Close()
	CreateTables(db)

	LogSession(db, 300, 5, 2)

	sessions, err := GetRecentSessions(db, 7)
	if err != nil {
		t.Fatalf("GetRecentSessions failed: %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("expected 1 session group, got %d", len(sessions))
	}
	if sessions[0].Reviewed != 5 {
		t.Errorf("expected 5 reviewed, got %d", sessions[0].Reviewed)
	}
}
