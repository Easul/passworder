package main

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"passworder/internal/config"
	"passworder/internal/database"
	"passworder/internal/handler"
	"passworder/internal/repository"
	"passworder/internal/service"
	"passworder/internal/storage"
)

func buildCLIOverrides(f serverFlags) config.CLIOverrides {
	return config.CLIOverrides{
		Host:                  f.host,
		Port:                  f.port,
		DBPath:                f.dbPath,
		StorageDir:            f.storageDir,
		ReminderCheckInterval: f.reminderInterval,
	}
}

func openDatabase(cfg *config.Config) (*sqlx.DB, error) {
	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0755); err != nil {
		return nil, err
	}
	db, err := database.Open(cfg.DBPath)
	if err != nil {
		return nil, err
	}
	if err := database.Migrate(db, migrationsFS); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func loadRuntimeConfig(db *sqlx.DB, initialCfg *config.Config, overrides config.CLIOverrides) *config.Config {
	settingService := service.NewSettingService(repository.NewSettingRepository(db))
	return applyDatabaseConfig(initialCfg, settingService, overrides)
}

func createServerRouter(cfg *config.Config, db *sqlx.DB, staticFS embed.FS) (http.Handler, error) {
	fileStore := storage.NewFileStorage(cfg.StorageDir)
	if err := fileStore.EnsureRoot(); err != nil {
		return nil, err
	}
	return handler.NewRouter(cfg, db, fileStore, staticFS), nil
}

func serverAddress(cfg *config.Config) string {
	return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
}
