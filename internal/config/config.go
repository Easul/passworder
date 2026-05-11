package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	DefaultHost                  = "127.0.0.1"
	DefaultPort                  = 18080
	DefaultDBName                = "password.db"
	DefaultStorageName           = "storage"
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

func getExecutableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func isRunningFromGoRun() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	exe = strings.ToLower(exe)
	return strings.HasPrefix(filepath.Base(exe), "__debug_bin") ||
		strings.HasPrefix(filepath.Base(exe), "go-build")
}

func GetDataDir() string {
	if isRunningFromGoRun() {
		_, filename, _, ok := runtime.Caller(0)
		if ok {
			projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
			return projectRoot
		}
	}
	return getExecutableDir()
}

func resolveDataPath(userPath, defaultName string) string {
	if userPath != "" {
		return userPath
	}
	return filepath.Join(GetDataDir(), defaultName)
}

func Load(overrides CLIOverrides) (*Config, error) {
	cfg := &Config{
		Host:                  DefaultHost,
		Port:                  DefaultPort,
		DBPath:                resolveDataPath(overrides.DBPath, DefaultDBName),
		StorageDir:            resolveDataPath(overrides.StorageDir, DefaultStorageName),
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
		DBPath:                resolveDataPath(overrides.DBPath, DefaultDBName),
		StorageDir:            resolveDataPath(overrides.StorageDir, DefaultStorageName),
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

	return cfg
}
