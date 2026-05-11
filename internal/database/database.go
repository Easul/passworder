package database

import (
	"fmt"
	"io/fs"
	"runtime"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

func Open(dbPath string) (*sqlx.DB, error) {
	driver := "sqlite3"
	if runtime.GOOS == "windows" {
		driver = "sqlite"
	}
	db, err := sqlx.Connect(driver, dbPath+"?_fk=1&busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("connect to sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	return db, nil
}

func Migrate(db *sqlx.DB, migrationsFS fs.FS) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at INTEGER DEFAULT (strftime('%s', 'now'))
	)`); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)

	for _, file := range files {
		var count int
		if err := db.Get(&count, `SELECT COUNT(1) FROM schema_migrations WHERE version = ?`, file); err != nil {
			return fmt.Errorf("check migration %s: %w", file, err)
		}
		if count > 0 {
			continue
		}

		content, err := fs.ReadFile(migrationsFS, "migrations/"+file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}

		tx, err := db.Beginx()
		if err != nil {
			return fmt.Errorf("begin transaction for %s: %w", file, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", file, err)
		}

		if _, err := tx.Exec(`INSERT INTO schema_migrations(version) VALUES(?)`, file); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", file, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", file, err)
		}
	}

	return nil
}
