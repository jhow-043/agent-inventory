// Package config handles the agent configuration loaded from a JSON file.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config holds the agent runtime configuration.
type Config struct {
	ServerURL          string        `json:"server_url"`
	EnrollmentKey      string        `json:"enrollment_key"`
	IntervalHours      int           `json:"interval_hours"`
	DataDir            string        `json:"data_dir"`
	LogLevel           string        `json:"log_level"`
	InsecureSkipVerify bool          `json:"insecure_skip_verify"`
	Interval           time.Duration `json:"-"`
}

// Load reads and parses the configuration from a JSON file.
// If path is empty, it looks for config.json next to the executable.
func Load(path string) (*Config, error) {
	if path == "" {
		exe, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("get executable path: %w", err)
		}
		path = filepath.Join(filepath.Dir(exe), "config.json")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	cfg.Interval = time.Duration(cfg.IntervalHours) * time.Hour

	if cfg.DataDir == "" {
		exe, _ := os.Executable()
		cfg.DataDir = filepath.Join(filepath.Dir(exe), "data")
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.ServerURL == "" {
		return fmt.Errorf("server_url is required")
	}
	if c.EnrollmentKey == "" {
		return fmt.Errorf("enrollment_key is required")
	}
	if c.IntervalHours <= 0 {
		c.IntervalHours = 1
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	return nil
}
