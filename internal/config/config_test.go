package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load returned nil config")
	}
}

func TestLoadConfigCreatesDefault(t *testing.T) {
	tmpDir := t.TempDir()
	origConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origConfigHome)

	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	origDataHome := os.Getenv("XDG_DATA_HOME")
	defer os.Setenv("XDG_DATA_HOME", origDataHome)
	os.Setenv("XDG_DATA_HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	configPath := filepath.Join(tmpDir, "irw", "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not auto-created")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var parsed map[string]string
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal("auto-created config is not valid JSON")
	}

	if dbPath := cfg.GetDBPath(); dbPath == "" {
		t.Error("GetDBPath returned empty after Load")
	}

	if launcher := cfg.GetLauncher(); launcher == "" {
		t.Error("GetLauncher returned empty after Load")
	}

	if cfg.DBPath == "" {
		t.Error("DBPath should be non-empty in auto-created config")
	}

	if !cfg.DBPathExplicit {
		t.Error("DBPathExplicit should be true in auto-created config")
	}

	_ = parsed
}

func TestLoadConfigPrePopulatedPath(t *testing.T) {
	tmpDir := t.TempDir()
	origXDG := os.Getenv("XDG_DATA_HOME")
	defer os.Setenv("XDG_DATA_HOME", origXDG)
	os.Setenv("XDG_DATA_HOME", tmpDir)

	origConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origConfigHome)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "irw-tool", "irw.db")
	if cfg.DBPath != expected {
		t.Errorf("expected db_path %q, got %q", expected, cfg.DBPath)
	}

	if !cfg.DBPathExplicit {
		t.Error("DBPathExplicit should be true for pre-populated config")
	}
}

func TestConfigProviderInterface(t *testing.T) {
	cfg, _ := Load()
	var _ ConfigProvider = cfg
}

func TestGetDBPathWithEmptyConfig(t *testing.T) {
	cfg := &Config{DBPath: "", Launcher: ""}
	path := cfg.GetDBPath()
	if path == "" {
		t.Error("GetDBPath should return default path, not empty")
	}
}

func TestGetLauncherWithEmptyConfig(t *testing.T) {
	cfg := &Config{DBPath: "", Launcher: ""}
	launcher := cfg.GetLauncher()
	if launcher == "" {
		t.Error("GetLauncher should return default launcher, not empty")
	}
}

func TestGetDBPathWithCustomConfig(t *testing.T) {
	cfg := &Config{DBPath: "/custom/path/db.db", Launcher: ""}
	if got := cfg.GetDBPath(); got != "/custom/path/db.db" {
		t.Errorf("expected /custom/path/db.db, got %s", got)
	}
}

func TestGetLauncherWithCustomConfig(t *testing.T) {
	cfg := &Config{DBPath: "", Launcher: "/usr/bin/zathura"}
	if got := cfg.GetLauncher(); got != "/usr/bin/zathura" {
		t.Errorf("expected /usr/bin/zathura, got %s", got)
	}
}

func TestNullConfig(t *testing.T) {
	var nc NullConfig
	if nc.GetDBPath() != "" {
		t.Error("NullConfig GetDBPath should return empty string")
	}
	if nc.GetLauncher() != "" {
		t.Error("NullConfig GetLauncher should return empty string")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "irw", "config.json")
	os.MkdirAll(filepath.Dir(configPath), 0755)
	os.WriteFile(configPath, []byte("invalid json{"), 0644)

	origConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origConfigHome)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLoadValidJSONWithMissingKeys(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "irw", "config.json")
	os.MkdirAll(filepath.Dir(configPath), 0755)
	os.WriteFile(configPath, []byte(`{"other_key": "value"}`), 0644)

	origConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origConfigHome)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.GetDBPath() == "" {
		t.Error("GetDBPath should return default when key missing")
	}
}