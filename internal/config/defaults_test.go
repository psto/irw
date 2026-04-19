package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefaultDBPath(t *testing.T) {
	path := DefaultDBPath()
	if path == "" {
		t.Fatal("DefaultDBPath returned empty")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("DefaultDBPath should return absolute path, got %s", path)
	}
	if filepath.Ext(filepath.Ext(path)) != ".db" {
		t.Errorf("DefaultDBPath should end with .db, got %s", path)
	}
}

func TestDefaultLauncher(t *testing.T) {
	launcher := DefaultLauncher()
	if launcher == "" {
		t.Fatal("DefaultLauncher returned empty")
	}
	valid := map[string]bool{"xdg-open": true, "open": true, "start": true}
	if !valid[launcher] {
		t.Errorf("unexpected launcher %q, expected one of: xdg-open, open, start", launcher)
	}
}

func TestConfigDir(t *testing.T) {
	dir := ConfigDir()
	if dir == "" {
		t.Fatal("ConfigDir returned empty")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("ConfigDir should return absolute path, got %s", dir)
	}
	expected := "irw"
	if filepath.Base(dir) != expected {
		t.Errorf("ConfigDir should end with %s, got %s", expected, dir)
	}
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()
	if path == "" {
		t.Fatal("ConfigPath returned empty")
	}
	if filepath.Ext(path) != ".json" {
		t.Errorf("ConfigPath should end with .json, got %s", path)
	}
	if filepath.Base(filepath.Dir(path)) != "irw" {
		t.Errorf("ConfigPath should be in irw directory, got %s", path)
	}
}

func TestDefaultDBPathDependsOnDataDir(t *testing.T) {
	origXDG := os.Getenv("XDG_DATA_HOME")
	defer os.Setenv("XDG_DATA_HOME", origXDG)

	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", tmpDir)

	path := DefaultDBPath()
	expected := filepath.Join(tmpDir, "irw-tool", "irw.db")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestDefaultLauncherLinux(t *testing.T) {
	if runtime.GOOS == "linux" {
		if launcher := DefaultLauncher(); launcher != "xdg-open" {
			t.Errorf("expected xdg-open on linux, got %s", launcher)
		}
	}
}

func TestDefaultLauncherDarwin(t *testing.T) {
	if runtime.GOOS == "darwin" {
		if launcher := DefaultLauncher(); launcher != "open" {
			t.Errorf("expected open on darwin, got %s", launcher)
		}
	}
}

func TestDefaultLauncherWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		if launcher := DefaultLauncher(); launcher != "start" {
			t.Errorf("expected start on windows, got %s", launcher)
		}
	}
}

