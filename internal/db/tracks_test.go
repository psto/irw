package db

import (
	"testing"
)

func TestInsertTrack(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := Connect(tmpDir + "/test.db")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer db.Close()
	CreateTables(db)

	err = InsertTrack(db, "/test/path.pdf", "reading", nil)
	if err != nil {
		t.Fatalf("InsertTrack failed: %v", err)
	}
}

func TestDeleteTrack(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := Connect(tmpDir + "/test.db")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer db.Close()
	CreateTables(db)

	InsertTrack(db, "/test/path.pdf", "reading", nil)
	err = DeleteTrack(db, "/test/path.pdf")
	if err != nil {
		t.Fatalf("DeleteTrack failed: %v", err)
	}
}

func TestMarkFinished(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := Connect(tmpDir + "/test.db")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer db.Close()
	CreateTables(db)

	InsertTrack(db, "/test/path.pdf", "reading", nil)
	err = MarkFinished(db, "/test/path.pdf")
	if err != nil {
		t.Fatalf("MarkFinished failed: %v", err)
	}
}

func TestUpdatePriority(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := Connect(tmpDir + "/test.db")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer db.Close()
	CreateTables(db)

	InsertTrack(db, "/test/path.pdf", "reading", nil)
	err = UpdatePriority(db, "/test/path.pdf", 75)
	if err != nil {
		t.Fatalf("UpdatePriority failed: %v", err)
	}
}

func TestUpdateInterval(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := Connect(tmpDir + "/test.db")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer db.Close()
	CreateTables(db)

	InsertTrack(db, "/test/path.pdf", "reading", nil)
	err = UpdateInterval(db, "/test/path.pdf", 5.0, 1700000000)
	if err != nil {
		t.Fatalf("UpdateInterval failed: %v", err)
	}
}

func TestGetNextDue(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := Connect(tmpDir + "/test.db")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer db.Close()
	CreateTables(db)

	InsertTrack(db, "/test/path.pdf", "reading", nil)
	path, interval, _, err := GetNextDue(db, "reading", "")
	if err != nil {
		t.Fatalf("GetNextDue failed: %v", err)
	}
	if path == "" {
		t.Error("expected a path, got empty")
	}
	if interval <= 0 {
		t.Errorf("expected positive interval, got %f", interval)
	}
}

func TestGetAllPaths(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := Connect(tmpDir + "/test.db")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer db.Close()
	CreateTables(db)

	InsertTrack(db, "/test/a.pdf", "reading", nil)
	InsertTrack(db, "/test/b.pdf", "reading", nil)

	paths, err := GetAllPaths(db)
	if err != nil {
		t.Fatalf("GetAllPaths failed: %v", err)
	}
	if len(paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(paths))
	}
}
