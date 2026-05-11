package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type serverFlags struct {
	host             string
	port             int
	dbPath           string
	storageDir       string
	reminderInterval int
	resetPassword    bool
	showVersion      bool
}

func parseServerFlags() serverFlags {
	var f serverFlags

	flag.StringVar(&f.host, "host", "", "HTTP server host (default: 0.0.0.0, env: PASSORDER_HOST)")
	flag.IntVar(&f.port, "port", 0, "HTTP server port (default: 8080, env: PASSORDER_PORT)")
	flag.StringVar(&f.dbPath, "db", "", "SQLite database path")
	flag.StringVar(&f.storageDir, "storage", "", "Storage directory path")
	flag.IntVar(&f.reminderInterval, "reminder-interval", 0, "Reminder check interval in minutes (default: 10)")
	flag.BoolVar(&f.resetPassword, "reset-password", false, "Reset master password")
	flag.BoolVar(&f.showVersion, "v", false, "Show version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: passworder [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -v                         Show version and exit\n")
		fmt.Fprintf(os.Stderr, "  -h, --help                 Show this help message\n")
		fmt.Fprintf(os.Stderr, "      --host <address>       HTTP server host (default: 0.0.0.0, env: PASSORDER_HOST)\n")
		fmt.Fprintf(os.Stderr, "      --port <number>        HTTP server port (default: 8080, env: PASSORDER_PORT)\n")
		fmt.Fprintf(os.Stderr, "      --db <path>            SQLite database path\n")
		fmt.Fprintf(os.Stderr, "      --storage <path>       Storage directory path\n")
		fmt.Fprintf(os.Stderr, "      --reminder-interval <minutes>  Reminder check interval (default: 10)\n")
		fmt.Fprintf(os.Stderr, "      --reset-password       Reset master password and exit\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  passworder --port 8080 --db ./password.db\n")
		fmt.Fprintf(os.Stderr, "  passworder --host 127.0.0.1 --port 9000\n")
		fmt.Fprintf(os.Stderr, "  passworder --reset-password  # Reset master password\n")
		fmt.Fprintf(os.Stderr, "  passworder -v  # Show version\n")
	}

	for _, arg := range os.Args[1:] {
		if arg == "--help" || arg == "-h" {
			flag.Usage()
			os.Exit(0)
		}
	}

	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") {
			flagName := strings.TrimPrefix(arg, "-")
			if idx := strings.Index(flagName, "="); idx > 0 {
				flagName = flagName[:idx]
			}
			if len(flagName) > 1 {
				fmt.Fprintf(os.Stderr, "Error: invalid flag format '%s'\n", arg)
				fmt.Fprintf(os.Stderr, "Long flags must use '--' prefix (e.g., --%s)\n", flagName)
				fmt.Fprintf(os.Stderr, "Short flags must be single character (e.g., -h)\n")
				fmt.Fprintf(os.Stderr, "\nUse -h or --help for usage information\n")
				os.Exit(1)
			}
			if len(flagName) > 0 && flagName[0] >= '0' && flagName[0] <= '9' {
				continue
			}
		}
	}

	flag.Parse()
	return f
}
