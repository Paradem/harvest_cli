package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	ProjectID          int64  `json:"project_id"`
	TaskID             int64  `json:"task_id"`
	HarvestAccountID   string `json:"harvest_account_id"`
	HarvestAccessToken string `json:"harvest_access_token"`
	HarvestUserID      string `json:"harvest_user_id"`
}

// DefaultConfigPath returns the default config file path (~/.harvestcli/config.json).
func DefaultConfigPath() string {
	return filepath.Join("./", ".harvestcli.json")
}

// GlobalConfigPath returns the global config file path (~/.config/harvest_cli/config.json).
func GlobalConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("./", ".harvestcli.json")
	}
	return filepath.Join(homeDir, ".config", "harvest_cli", "config.json")
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

// LoadGlobal loads the global config from ~/.config/harvest_cli/config.json.
func LoadGlobal() (*Config, error) {
	path := GlobalConfigPath()
	return Load(path)
}

// SaveGlobal saves the config to the global config path, creating directories as needed.
func (c *Config) SaveGlobal() error {
	path := GlobalConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return c.Save(path)
}

// SetupLogger creates a logger that writes to ~/.config/harvest_cli/debug.log
func SetupLogger() (*log.Logger, error) {
	logPath := filepath.Join(filepath.Dir(GlobalConfigPath()), "debug.log")

	// Create directory if it doesn't exist
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	// Open log file
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return nil, err
	}

	// Create logger
	logger := log.New(file, "", log.LstdFlags)
	return logger, nil
}
