package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type ConfigProvider interface {
	GetDBPath() string
	GetLauncher() string
}

type Config struct {
	DBPath         string `json:"db_path"`
	Launcher       string `json:"launcher"`
	DBPathExplicit bool   `json:"-"`
}

func (c *Config) UnmarshalJSON(data []byte) error {
	type aliasStruct struct {
		DBPath   string `json:"db_path"`
		Launcher string `json:"launcher"`
	}
	var tmp aliasStruct
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	c.DBPath = tmp.DBPath
	c.Launcher = tmp.Launcher
	c.DBPathExplicit = tmp.DBPath != ""
return nil
}

func (c *Config) GetDBPath() string {
	if c.DBPath != "" {
		return c.DBPath
	}
	return DefaultDBPath()
}

func (c *Config) GetLauncher() string {
	if c.Launcher != "" {
		return c.Launcher
	}
	return DefaultLauncher()
}

func Load() (*Config, error) {
	cfg := &Config{}

	configPath := ConfigPath()
	dir := ConfigDir()

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config dir %s: %w", dir, err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		if err := writeDefaultConfig(configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		data, err = os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read newly created config: %w", err)
		}
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}

func writeDefaultConfig(path string) error {
	cfg := &Config{
		DBPath:   DefaultDBPath(),
		Launcher: "",
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

type NullConfig struct{}

func (n NullConfig) GetDBPath() string {
	return ""
}

func (n NullConfig) GetLauncher() string {
	return ""
}