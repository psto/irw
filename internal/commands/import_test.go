package commands

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/psto/irw/internal/db"
	"github.com/psto/irw/internal/models"
	"github.com/psto/irw/internal/tui"
)

type mockZkRunner struct {
	notes map[string][]models.ZkNote
}

func (m mockZkRunner) Run(args ...string) ([]byte, error) {
	for _, arg := range args {
		if arg == "--tag" {
			tagIdx := 0
			for i, a := range args {
				if a == "--tag" {
					tagIdx = i + 1
					break
				}
			}
			if tagIdx < len(args) {
				tag := args[tagIdx]
				if notes, ok := m.notes[tag]; ok {
					return json.Marshal(notes)
				}
			}
		}
	}
	return json.Marshal([]models.ZkNote{})
}

type mockConfigProvider struct {
	zkTags map[string][]string
}

func (m mockConfigProvider) GetDBPath() string {
	return ""
}

func (m mockConfigProvider) GetLauncher() string {
	return ""
}

func (m mockConfigProvider) GetZkTags() map[string][]string {
	return m.zkTags
}

type mockFileExists struct {
	existingPaths map[string]bool
}

func (m mockFileExists) FileExists(path string) bool {
	return m.existingPaths[path]
}

func setupTestDB(t *testing.T) (*sql.DB, string) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	if err := db.CreateTables(database); err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}
	return database, tmpDir
}

func createMockFile(t *testing.T, tmpDir, filename string) string {
	absPath := filepath.Join(tmpDir, filename)
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if _, err := os.Create(absPath); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	return absPath
}

func TestImportCleanupWhenTagRemoved(t *testing.T) {
	database, tmpDir := setupTestDB(t)
	defer database.Close()

	article1 := createMockFile(t, tmpDir, "article1.md")
	article2 := createMockFile(t, tmpDir, "article2.md")

	cfg := mockConfigProvider{
		zkTags: map[string][]string{
			"reading": {"status/reading"},
		},
	}

	zk := mockZkRunner{
		notes: map[string][]models.ZkNote{
			"status/reading": {
				{AbsPath: article1},
				{AbsPath: article2},
			},
		},
	}

	importFromZk(database, cfg, zk)

	exists, _ := db.PathExists(database, article1)
	if !exists {
		t.Error("expected article1.md to exist after import")
	}

	cfg = mockConfigProvider{
		zkTags: map[string][]string{
			"reading": {},
		},
	}

	zk = mockZkRunner{
		notes: map[string][]models.ZkNote{
			"status/reading": {},
		},
	}

	importFromZk(database, cfg, zk)

	exists, _ = db.PathExists(database, article1)
	if exists {
		t.Error("expected article1.md to be deleted after tag removed from config")
	}
}

func TestImportPartialCleanupSomeTagsRemoved(t *testing.T) {
	database, tmpDir := setupTestDB(t)
	defer database.Close()

	article1 := createMockFile(t, tmpDir, "article1.md")
	research1 := createMockFile(t, tmpDir, "research1.md")

	cfg := mockConfigProvider{
		zkTags: map[string][]string{
			"reading": {"status/reading", "status/research"},
		},
	}

	zk := mockZkRunner{
		notes: map[string][]models.ZkNote{
			"status/reading": {
				{AbsPath: article1},
			},
			"status/research": {
				{AbsPath: research1},
			},
		},
	}

	importFromZk(database, cfg, zk)

	cfg = mockConfigProvider{
		zkTags: map[string][]string{
			"reading": {"status/reading"},
		},
	}

	zk = mockZkRunner{
		notes: map[string][]models.ZkNote{
			"status/reading": {
				{AbsPath: article1},
			},
		},
	}

	importFromZk(database, cfg, zk)

	exists1, _ := db.PathExists(database, article1)
	if !exists1 {
		t.Error("expected article1.md (status/reading) to remain")
	}

	exists2, _ := db.PathExists(database, research1)
	if exists2 {
		t.Error("expected research1.md (status/research) to be deleted")
	}
}

func TestImportMultipleTypesCleanupAffectsCorrectType(t *testing.T) {
	database, tmpDir := setupTestDB(t)
	defer database.Close()

	article1 := createMockFile(t, tmpDir, "article1.md")
	draft1 := createMockFile(t, tmpDir, "draft1.md")

	cfg := mockConfigProvider{
		zkTags: map[string][]string{
			"reading": {"status/reading"},
			"writing": {"status/writing"},
		},
	}

	zk := mockZkRunner{
		notes: map[string][]models.ZkNote{
			"status/reading": {
				{AbsPath: article1},
			},
			"status/writing": {
				{AbsPath: draft1},
			},
		},
	}

	importFromZk(database, cfg, zk)

	cfg = mockConfigProvider{
		zkTags: map[string][]string{
			"reading": {},
			"writing": {"status/writing"},
		},
	}

	zk = mockZkRunner{
		notes: map[string][]models.ZkNote{
			"status/reading": {},
			"status/writing": {
				{AbsPath: draft1},
			},
		},
	}

	importFromZk(database, cfg, zk)

	readingExists, _ := db.PathExists(database, article1)
	if readingExists {
		t.Error("expected reading track to be deleted")
	}

	writingExists, _ := db.PathExists(database, draft1)
	if !writingExists {
		t.Error("expected writing track to remain")
	}
}

func TestImportNoCleanupWhenNoTagsRemoved(t *testing.T) {
	database, tmpDir := setupTestDB(t)
	defer database.Close()

	article1 := createMockFile(t, tmpDir, "article1.md")

	cfg := mockConfigProvider{
		zkTags: map[string][]string{
			"reading": {"status/reading"},
		},
	}

	zk := mockZkRunner{
		notes: map[string][]models.ZkNote{
			"status/reading": {
				{AbsPath: article1},
			},
		},
	}

	importFromZk(database, cfg, zk)

	importFromZk(database, cfg, zk)

	exists, _ := db.PathExists(database, article1)
	if !exists {
		t.Error("expected article1.md to remain after second import (idempotent)")
	}
}

func TestImportLeavesManualTracks(t *testing.T) {
	database, tmpDir := setupTestDB(t)
	defer database.Close()

	article1 := createMockFile(t, tmpDir, "article1.md")
	manualNote := createMockFile(t, tmpDir, "manual_note.md")

	cfg := mockConfigProvider{
		zkTags: map[string][]string{
			"reading": {"status/reading"},
		},
	}

	zk := mockZkRunner{
		notes: map[string][]models.ZkNote{
			"status/reading": {
				{AbsPath: article1},
			},
		},
	}

	importFromZk(database, cfg, zk)

	db.InsertTrack(database, manualNote, "reading", nil)

	cfg = mockConfigProvider{
		zkTags: map[string][]string{
			"reading": {},
		},
	}

	zk = mockZkRunner{
		notes: map[string][]models.ZkNote{
			"status/reading": {},
		},
	}

	importFromZk(database, cfg, zk)

	zkExists, _ := db.PathExists(database, article1)
	if zkExists {
		t.Error("expected zk track to be deleted")
	}
}

func TestImportCleansUpForConfiguredTypesOnly(t *testing.T) {
	database, tmpDir := setupTestDB(t)
	defer database.Close()

	article1 := createMockFile(t, tmpDir, "article1.md")
	otherPath := createMockFile(t, tmpDir, "other.md")

	cfg := mockConfigProvider{
		zkTags: map[string][]string{
			"reading": {"status/reading"},
		},
	}

	zk := mockZkRunner{
		notes: map[string][]models.ZkNote{
			"status/reading": {
				{AbsPath: article1},
			},
		},
	}

	importFromZk(database, cfg, zk)

	db.InsertTrack(database, otherPath, "other_type", nil)

	cfg = mockConfigProvider{
		zkTags: map[string][]string{
			"reading": {},
		},
	}

	zk = mockZkRunner{
		notes: map[string][]models.ZkNote{
			"status/reading": {},
		},
	}

	importFromZk(database, cfg, zk)

	otherExists, _ := db.PathExists(database, otherPath)
	if !otherExists {
		t.Error("expected other_type track to remain (not in configured types)")
	}
}

var _ tui.ZkRunner = mockZkRunner{}