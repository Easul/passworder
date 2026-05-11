package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"

	"passworder/internal/config"
)

//go:embed static/*
//go:embed static/css/*
//go:embed static/js/*
var staticFS embed.FS

//go:embed all:migrations/*.sql
var migrationsFS embed.FS

const Version = "v1.0.0"

func main() {
	flags := parseServerFlags()
	if flags.showVersion {
		fmt.Println(Version)
		os.Exit(0)
	}

	overrides := buildCLIOverrides(flags)
	initialCfg := config.LoadFromEnvAndDefaults(overrides)

	if flags.resetPassword {
		if err := resetMasterPassword(initialCfg.DBPath); err != nil {
			log.Fatalf("Failed to reset password: %v", err)
		}
		os.Exit(0)
	}

	db, err := openDatabase(initialCfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	cfg := loadRuntimeConfig(db, initialCfg, overrides)
	router, err := createServerRouter(cfg, db, staticFS)
	if err != nil {
		log.Fatalf("Failed to create server router: %v", err)
	}

	startAutoReminder(db, cfg.ReminderCheckInterval)

	addr := serverAddress(cfg)
	log.Printf("Server starting on http://%s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
