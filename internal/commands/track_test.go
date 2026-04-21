package commands

import (
	"sort"
	"strings"
	"testing"

	"github.com/psto/irw/internal/config"
)

type trackMockConfig struct {
	defaultQueue string
	zkTags       map[string][]string
}

func (m trackMockConfig) GetDBPath() string {
	return ""
}

func (m trackMockConfig) GetLauncher() string {
	return ""
}

func (m trackMockConfig) GetDefaultQueue() string {
	return m.defaultQueue
}

func (m trackMockConfig) GetZkTags() map[string][]string {
	return m.zkTags
}

var _ config.ConfigProvider = trackMockConfig{}

func TestTrackWithQueueFlag(t *testing.T) {
	cfg := trackMockConfig{
		defaultQueue: "reading",
		zkTags: map[string][]string{
			"reading": {"status/reading"},
			"writing": {"status/writing"},
		},
	}

	tests := []struct {
		name      string
		queueType string
		wantErr   bool
		wantType  string
		isURI     bool
	}{
		{
			name:      "reading queue",
			queueType: "reading",
			wantErr:   false,
			wantType:  "reading",
		},
		{
			name:      "writing queue",
			queueType: "writing",
			wantErr:   false,
			wantType:  "writing",
		},
		{
			name:      "invalid queue type",
			queueType: "invalid",
			wantErr:   true,
		},
		{
			name:      "empty queue defaults to reading",
			queueType: "",
			wantErr:   false,
			wantType:  "reading",
		},
		{
			name:      "URI with reading queue",
			queueType: "reading",
			wantErr:   false,
			wantType:  "reading",
			isURI:     true,
		},
		{
			name:      "URI with writing queue",
			queueType: "writing",
			wantErr:   false,
			wantType:  "writing",
			isURI:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database, tmpDir := setupTestDB(t)
			defer database.Close()

			var input string
			if tt.isURI {
				input = "https://example.com/article"
			} else {
				input = createMockFile(t, tmpDir, "test.pdf")
			}

			err := Track(cfg, database, input, tt.queueType)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var trackType string
			err = database.QueryRow("SELECT type FROM tracks WHERE path = ?", input).Scan(&trackType)
			if err != nil {
				t.Fatalf("failed to query track type: %v", err)
			}
			if trackType != tt.wantType {
				t.Errorf("expected type %q, got %q", tt.wantType, trackType)
			}
		})
	}
}

func TestTrackCustomQueue(t *testing.T) {
	cfg := trackMockConfig{
		defaultQueue: "reading",
		zkTags: map[string][]string{
			"reading":  {"status/reading"},
			"writing":  {"status/writing"},
			"research": {"status/research"},
		},
	}

	database, tmpDir := setupTestDB(t)
	defer database.Close()

	testFile := createMockFile(t, tmpDir, "paper.pdf")

	err := Track(cfg, database, testFile, "research")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var trackType string
	err = database.QueryRow("SELECT type FROM tracks WHERE path = ?", testFile).Scan(&trackType)
	if err != nil {
		t.Fatalf("failed to query track type: %v", err)
	}
	if trackType != "research" {
		t.Errorf("expected type %q, got %q", "research", trackType)
	}
}

func TestTrackCustomDefaultQueue(t *testing.T) {
	cfg := trackMockConfig{
		defaultQueue: "writing",
		zkTags: map[string][]string{
			"reading": {"status/reading"},
			"writing": {"status/writing"},
		},
	}

	database, tmpDir := setupTestDB(t)
	defer database.Close()

	testFile := createMockFile(t, tmpDir, "draft.md")

	err := Track(cfg, database, testFile, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var trackType string
	err = database.QueryRow("SELECT type FROM tracks WHERE path = ?", testFile).Scan(&trackType)
	if err != nil {
		t.Fatalf("failed to query track type: %v", err)
	}
	if trackType != "writing" {
		t.Errorf("expected type %q (default), got %q", "writing", trackType)
	}
}

func TestTrackInvalidQueueListsAllQueues(t *testing.T) {
	cfg := trackMockConfig{
		defaultQueue: "reading",
		zkTags: map[string][]string{
			"reading":  {"status/reading"},
			"writing":  {"status/writing"},
			"research": {"status/research"},
		},
	}

	database, tmpDir := setupTestDB(t)
	defer database.Close()

	testFile := createMockFile(t, tmpDir, "test.pdf")

	err := Track(cfg, database, testFile, "foobar")
	if err == nil {
		t.Fatal("expected error for invalid queue type")
	}

	errMsg := err.Error()
	queues := []string{"reading", "research", "writing"}
	sort.Strings(queues)
	wantPart := "configured queues: " + strings.Join(queues, ", ")
	if !containsString(errMsg, wantPart) {
		t.Errorf("error message should contain %q, got %q", wantPart, errMsg)
	}
}

func TestTrackInvalidQueueNotInConfig(t *testing.T) {
	cfg := trackMockConfig{
		defaultQueue: "reading",
		zkTags: map[string][]string{
			"reading": {"status/reading"},
			"writing": {"status/writing"},
		},
	}

	database, tmpDir := setupTestDB(t)
	defer database.Close()

	testFile := createMockFile(t, tmpDir, "test.pdf")

	err := Track(cfg, database, testFile, "research")
	if err == nil {
		t.Fatal("expected error for queue not in config")
	}
}

func TestTrackNonExistentFileReturnsError(t *testing.T) {
	cfg := trackMockConfig{
		defaultQueue: "reading",
		zkTags: map[string][]string{
			"reading": {"status/reading"},
		},
	}

	database, _ := setupTestDB(t)
	defer database.Close()

	err := Track(cfg, database, "/nonexistent/path.pdf", "reading")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}