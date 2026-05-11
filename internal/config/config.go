package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultHost                  = "0.0.0.0"
	DefaultPort                  = 8080
	DefaultDBPath                = "./password.db"
	DefaultStorageDir            = "./storage"
	DefaultReminderCheckInterval = 10
)

type Config struct {
	Host                  string `yaml:"host"`
	Port                  int    `yaml:"port"`
	DBPath                string `yaml:"db_path"`
	StorageDir            string `yaml:"storage_dir"`
	ReminderCheckInterval int    `yaml:"reminder_check_interval"`
}

type CLIOverrides struct {
	Host                  string
	Port                  int
	DBPath                string
	StorageDir            string
	ReminderCheckInterval int
}

func Load(overrides CLIOverrides) (*Config, error) {
	cfg := &Config{
		Host:                  DefaultHost,
		Port:                  DefaultPort,
		DBPath:                DefaultDBPath,
		StorageDir:            DefaultStorageDir,
		ReminderCheckInterval: DefaultReminderCheckInterval,
	}

	if interval := os.Getenv("PASSORDER_REMINDER_INTERVAL"); interval != "" {
		fmt.Sscanf(interval, "%d", &cfg.ReminderCheckInterval)
	}
	if overrides.ReminderCheckInterval > 0 {
		cfg.ReminderCheckInterval = overrides.ReminderCheckInterval
	}

	if host := os.Getenv("PASSORDER_HOST"); host != "" {
		cfg.Host = host
	}
	if port := os.Getenv("PASSORDER_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Port)
	}

	if overrides.Host != "" {
		cfg.Host = overrides.Host
	}
	if overrides.Port > 0 {
		cfg.Port = overrides.Port
	}
	if overrides.DBPath != "" {
		cfg.DBPath = overrides.DBPath
	}
	if overrides.StorageDir != "" {
		cfg.StorageDir = overrides.StorageDir
	}

	if err := os.MkdirAll(cfg.StorageDir, 0755); err != nil {
		return nil, err
	}

	absDB, err := filepath.Abs(cfg.DBPath)
	if err != nil {
		return nil, err
	}
	cfg.DBPath = absDB

	absStorage, err := filepath.Abs(cfg.StorageDir)
	if err != nil {
		return nil, err
	}
	cfg.StorageDir = absStorage

	return cfg, nil
}

func LoadFromEnvAndDefaults(overrides CLIOverrides) *Config {
	cfg := &Config{
		Host:                  DefaultHost,
		Port:                  DefaultPort,
		DBPath:                DefaultDBPath,
		StorageDir:            DefaultStorageDir,
		ReminderCheckInterval: DefaultReminderCheckInterval,
	}

	if interval := os.Getenv("PASSORDER_REMINDER_INTERVAL"); interval != "" {
		fmt.Sscanf(interval, "%d", &cfg.ReminderCheckInterval)
	}
	if overrides.ReminderCheckInterval > 0 {
		cfg.ReminderCheckInterval = overrides.ReminderCheckInterval
	}

	if host := os.Getenv("PASSORDER_HOST"); host != "" {
		cfg.Host = host
	}
	if port := os.Getenv("PASSORDER_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Port)
	}

	if overrides.Host != "" {
		cfg.Host = overrides.Host
	}
	if overrides.Port > 0 {
		cfg.Port = overrides.Port
	}
	if overrides.DBPath != "" {
		cfg.DBPath = overrides.DBPath
	}
	if overrides.StorageDir != "" {
		cfg.StorageDir = overrides.StorageDir
	}

	return cfg
}
