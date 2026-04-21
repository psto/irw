package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type ConfigProvider interface {
	GetDBPath() string
	GetLauncher() string
	GetDefaultQueue() string
	GetZkTags() map[string][]string
}

type Config struct {
	DBPath        string            `json:"db_path"`
	Launcher      string            `json:"launcher"`
	DefaultQueue  string            `json:"default_queue"`
	ZkTags        map[string][]string `json:"zk_tags"`
	DBPathExplicit bool             `json:"-"`
}

func (c *Config) UnmarshalJSON(data []byte) error {
	type aliasStruct struct {
		DBPath       string            `json:"db_path"`
		Launcher     string            `json:"launcher"`
		DefaultQueue string            `json:"default_queue"`
		ZkTags       map[string][]string `json:"zk_tags"`
	}
	var tmp aliasStruct
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	c.DBPath = tmp.DBPath
	c.Launcher = tmp.Launcher
	c.DefaultQueue = tmp.DefaultQueue
	c.ZkTags = tmp.ZkTags
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

func (c *Config) GetDefaultQueue() string {
	if c.DefaultQueue != "" {
		return c.DefaultQueue
	}
	return "reading"
}

func (c *Config) GetZkTags() map[string][]string {
	if len(c.ZkTags) > 0 {
		return c.ZkTags
	}
	return map[string][]string{
		"reading": {"status/reading"},
		"writing": {"status/writing"},
	}
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
		DBPath:       DefaultDBPath(),
		Launcher:     "",
		DefaultQueue: "reading",
		ZkTags: map[string][]string{
			"reading": {"status/reading"},
			"writing": {"status/writing"},
		},
	}
data, err := json.MarshalIndent(cfg, "", " ")
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

func (n NullConfig) GetDefaultQueue() string {
	return "reading"
}

func (n NullConfig) GetZkTags() map[string][]string {
	return map[string][]string{
		"reading": {"status/reading"},
		"writing": {"status/writing"},
	}
}