package commands

import (
	"testing"
)

func TestTrackWithQueueFlag(t *testing.T) {
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

			err := Track(database, input, tt.queueType)

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

func TestTrackInvalidQueueReturnsHelpfulError(t *testing.T) {
	database, tmpDir := setupTestDB(t)
	defer database.Close()

	testFile := createMockFile(t, tmpDir, "test.pdf")

	err := Track(database, testFile, "foobar")
	if err == nil {
		t.Fatal("expected error for invalid queue type")
	}

	wantMsg := "must be 'reading' or 'writing'"
	if !containsString(err.Error(), wantMsg) {
		t.Errorf("error message should contain %q, got %q", wantMsg, err.Error())
	}
}

func TestTrackNonExistentFileReturnsError(t *testing.T) {
	database, _ := setupTestDB(t)
	defer database.Close()

	err := Track(database, "/nonexistent/path.pdf", "reading")
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