package config

import (
	"os"
	"path/filepath"
	"runtime"
)

func userConfigDir() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	if home == "" {
		return ""
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support")
	case "windows":
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return appdata
		}
		return filepath.Join(home, "AppData", "Roaming")
	default:
		return filepath.Join(home, ".config")
	}
}

func userDataDir() string {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	if home == "" {
		return ""
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support")
	case "windows":
		if localappdata := os.Getenv("LOCALAPPDATA"); localappdata != "" {
			return localappdata
		}
		return filepath.Join(home, "AppData", "Local")
	default:
		return filepath.Join(home, ".local", "share")
	}
}

func DefaultDBPath() string {
	dir := userDataDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "irw-tool", "irw.db")
}

func DefaultLauncher() string {
	switch runtime.GOOS {
	case "windows":
		return "start"
	case "darwin":
		return "open"
	default:
		return "xdg-open"
	}
}

func ConfigDir() string {
	return filepath.Join(userConfigDir(), "irw")
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

