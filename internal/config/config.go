package config

import (
	"encoding/json"

	"os"
	"path/filepath"
)

type Config struct {
	ProjectID int64 `json:"project_id"`
	TaskID    int64 `json:"task_id"`
}

// DefaultConfigPath returns the default config file path (~/.harvestcli/config.json).
func DefaultConfigPath() string {
	return filepath.Join("./", ".harvestcli.json")
}

// Load reads the config file if it exists. If not, returns empty Config and nil error.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the config to disk.
func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
