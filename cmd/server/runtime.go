package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/term"
	"passworder/internal/auth"
	"passworder/internal/config"
	"passworder/internal/database"
	"passworder/internal/model"
	"passworder/internal/repository"
	"passworder/internal/service"
)

func applyDatabaseConfig(cfg *config.Config, settingService *service.SettingService, cli config.CLIOverrides) *config.Config {
	dbConfig := settingService.GetServerConfig()

	if cli.Host == "" && dbConfig.Host != "" {
		cfg.Host = dbConfig.Host
	}
	if cli.Port == 0 && dbConfig.Port > 0 {
		cfg.Port = dbConfig.Port
	}
	if cli.StorageDir == "" && dbConfig.StorageDir != "" {
		cfg.StorageDir = dbConfig.StorageDir
	}
	if cli.ReminderCheckInterval == 0 && dbConfig.ReminderCheckInterval > 0 {
		cfg.ReminderCheckInterval = dbConfig.ReminderCheckInterval
	}

	return cfg
}

func resetMasterPassword(dbPath string) error {
	fmt.Println("Resetting master password...")

	db, err := database.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	if err := database.Migrate(db, migrationsFS); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	authRepo := repository.NewAuthRepository(db)
	authService := auth.NewService()

	fmt.Print("Enter new master password: ")
	newPassword, err := readPassword()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	if len(newPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	fmt.Print("Confirm new master password: ")
	confirmPassword, err := readPassword()
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	if newPassword != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	passwordHash, salt, err := authService.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	authRecord := &model.Auth{
		PasswordHash: string(passwordHash),
		KDFSalt:      salt,
		CreatedAt:    model.Now(),
		UpdatedAt:    model.Now(),
	}
	if err := authRepo.Save(authRecord); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	fmt.Println("Master password has been reset successfully.")
	return nil
}

func readPassword() (string, error) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", err
		}
		fmt.Println()
		return string(bytePassword), nil
	}

	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(password), nil
}

func startAutoReminder(db *sqlx.DB, intervalMinutes int) {
	reminderRepo := repository.NewReminderRepository(db)
	settingService := service.NewSettingService(repository.NewSettingRepository(db))
	mailSender := service.NewSMTPSender()
	reminderService := service.NewReminderService(reminderRepo, settingService, mailSender)

	if intervalMinutes <= 0 {
		intervalMinutes = 10
	}

	go func() {
		log.Printf("Auto-reminder service started (immediate check + every %d minutes)", intervalMinutes)

		log.Println("Auto-reminder: checking for due reminders on startup...")
		result, err := reminderService.SendDue()
		if err != nil {
			log.Printf("Auto-reminder: initial check failed: %v", err)
		} else if result["sentGroups"] > 0 {
			log.Printf("Auto-reminder: startup sent %d groups, %d reminders", result["sentGroups"], result["sentReminders"])
		} else {
			log.Println("Auto-reminder: no due reminders on startup")
		}

		ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			result, err := reminderService.SendDue()
			if err != nil {
				log.Printf("Auto-reminder: failed to send: %v", err)
				continue
			}
			if result["sentGroups"] > 0 {
				log.Printf("Auto-reminder: sent %d groups, %d reminders", result["sentGroups"], result["sentReminders"])
			}
		}
	}()
}
