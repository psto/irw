package commands

import (
	"strings"
	"testing"

	"github.com/psto/irw/internal/db"
)

type readMockConfig struct {
	launcher string
}

func (m readMockConfig) GetDBPath() string        { return "" }
func (m readMockConfig) GetLauncher() string      { return m.launcher }
func (m readMockConfig) GetDefaultQueue() string  { return "reading" }
func (m readMockConfig) GetZkTags() map[string][]string {
	return map[string][]string{"reading": {"status/reading"}}
}

func TestReadFileWithUntrackedPath(t *testing.T) {
	database, tmpDir := setupTestDB(t)
	defer database.Close()

	untrackedFile := createMockFile(t, tmpDir, "untracked.pdf")
	cfg := readMockConfig{launcher: "echo"}

	err := ReadFile(cfg, database, untrackedFile)
	if err == nil {
		t.Fatal("expected error for untracked path, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "not tracked") {
		t.Errorf("error should mention 'not tracked', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "irw track") {
		t.Errorf("error should suggest 'irw track', got: %s", errMsg)
	}
}

func TestReadFileWithUntrackedAbsolutePath(t *testing.T) {
	database, _ := setupTestDB(t)
	defer database.Close()

	cfg := readMockConfig{launcher: "echo"}

	err := ReadFile(cfg, database, "/absolutely/not/tracked/file.pdf")
	if err == nil {
		t.Fatal("expected error for untracked absolute path, got nil")
	}

	if !strings.Contains(err.Error(), "irw track") {
		t.Errorf("error should suggest 'irw track', got: %s", err.Error())
	}
}

func TestReadFileNoArgsEmptyDB(t *testing.T) {
	database, _ := setupTestDB(t)
	defer database.Close()

	cfg := readMockConfig{launcher: "echo"}

	err := ReadFile(cfg, database, "")
	if err != nil {
		t.Fatalf("expected no error for empty DB, got: %v", err)
	}

	var count int
	database.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 sessions for empty DB, got %d", count)
	}
}

func TestReadFileWithTrackedPath(t *testing.T) {
	database, tmpDir := setupTestDB(t)
	defer database.Close()

	testFile := createMockFile(t, tmpDir, "article.pdf")
	trackCfg := trackMockConfig{defaultQueue: "reading", zkTags: map[string][]string{"reading": {"status/reading"}}}

	if err := Track(trackCfg, database, testFile, "reading"); err != nil {
		t.Fatalf("failed to track file: %v", err)
	}

	readCfg := readMockConfig{launcher: "true"}

	err := ReadFile(readCfg, database, testFile)
	if err != nil {
		t.Fatalf("expected no error for tracked path, got: %v", err)
	}

	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM sessions WHERE reviewed = 1").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query sessions: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 session logged, got %d", count)
	}
}

func TestPathExistsInDB(t *testing.T) {
	database, tmpDir := setupTestDB(t)
	defer database.Close()

	testFile := createMockFile(t, tmpDir, "exists.pdf")
	trackCfg := trackMockConfig{defaultQueue: "reading", zkTags: map[string][]string{"reading": {"status/reading"}}}

	Track(trackCfg, database, testFile, "reading")

	exists, err := db.PathExists(database, testFile)
	if err != nil {
		t.Fatalf("PathExists error: %v", err)
	}
	if !exists {
		t.Error("expected tracked file to exist in DB")
	}

	exists, err = db.PathExists(database, "/no/such/file.pdf")
	if err != nil {
		t.Fatalf("PathExists error: %v", err)
	}
	if exists {
		t.Error("expected untracked file to not exist in DB")
	}
}
