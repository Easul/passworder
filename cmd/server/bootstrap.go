package main

import "passworder/internal/config"

func buildCLIOverrides(f serverFlags) config.CLIOverrides {
	return config.CLIOverrides{
		Host:                  f.host,
		Port:                  f.port,
		DBPath:                f.dbPath,
		StorageDir:            f.storageDir,
		ReminderCheckInterval: f.reminderInterval,
	}
}
