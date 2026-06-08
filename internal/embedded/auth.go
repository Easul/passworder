package embedded

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
	"passworder/internal/auth"
	"passworder/internal/database"
	"passworder/internal/model"
	"passworder/internal/repository"
)

func ResetMasterPassword(dbPath string) error {
	db, err := database.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	if err := database.Migrate(db, embeddedAssetsFS()); err != nil {
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
