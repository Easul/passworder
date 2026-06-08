package main

import (
	"fmt"
	"log"
	"os"

	"passworder/internal/config"
	"passworder/internal/embedded"
)

func main() {
	flags := parseServerFlags()
	if flags.showVersion {
		fmt.Println(embedded.Version)
		os.Exit(0)
	}

	overrides := buildCLIOverrides(flags)
	initialCfg := config.LoadFromEnvAndDefaults(overrides)

	if flags.resetPassword {
		if err := embedded.ResetMasterPassword(initialCfg.DBPath); err != nil {
			log.Fatalf("Failed to reset password: %v", err)
		}
		os.Exit(0)
	}

	server, err := embedded.NewEmbeddedServer(overrides)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer server.Stop()
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", server.Config().Host, server.Config().Port)
	log.Printf("Server starting on http://%s", addr)
	if err := server.Wait(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
