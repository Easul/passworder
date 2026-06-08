package main

import (
	"passworder/internal/embedded"
)

func resetMasterPassword(dbPath string) error {
	return embedded.ResetMasterPassword(dbPath)
}
