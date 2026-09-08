// Package store owns Cassor's SQLite persistence and database initialization.
package store

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/rickcern44/cassor/internal/migrations"
)

const DatabaseFileName = "cassor.db"

var defaultCategories = []string{
	"Needed",
	"Recommended",
	"Stretch goal",
	"Technical debt",
	"Security",
	"Research",
}

// Initialize creates or opens the database at path, applies all embedded
// migrations, and installs Cassor's initial categories.
func Initialize(path string) error {
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	if _, err := database.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}
	if err := applyMigrations(database); err != nil {
		return err
	}
	if err := seedCategories(database); err != nil {
		return err
	}
	return nil
}

// Migrate applies newly embedded schema migrations to an existing database.
func Migrate(path string) error {
	database, err := Open(path)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()
	database.SetMaxOpenConns(1)
	return applyMigrations(database)
}

func applyMigrations(database *sql.DB) error {
	if _, err := database.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		return fmt.Errorf("create migration metadata: %w", err)
	}
	names, err := migrations.Names()
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}

	for _, name := range names {
		version := strings.TrimSuffix(filepath.Base(name), ".sql")
		var exists bool
		if err := database.QueryRow(`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)`, version).Scan(&exists); err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if exists {
			continue
		}

		source, err := migrations.Files.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		disableForeignKeys := version == "0005_conventional_statuses"
		if disableForeignKeys {
			if _, err := database.Exec(`PRAGMA foreign_keys = OFF`); err != nil {
				return fmt.Errorf("disable foreign keys for migration %s: %w", version, err)
			}
		}
		transaction, err := database.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", version, err)
		}
		if _, err := transaction.Exec(string(source)); err != nil {
			transaction.Rollback()
			return fmt.Errorf("apply migration %s: %w", version, err)
		}
		if version == "0011_delivery_contract" {
			if err := backfillDeliveryContract(transaction); err != nil {
				transaction.Rollback()
				return fmt.Errorf("backfill migration %s: %w", version, err)
			}
		}
		if _, err := transaction.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, version, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			transaction.Rollback()
			return fmt.Errorf("record migration %s: %w", version, err)
		}
		if err := transaction.Commit(); err != nil {
			if disableForeignKeys {
				_, _ = database.Exec(`PRAGMA foreign_keys = ON`)
			}
			return fmt.Errorf("commit migration %s: %w", version, err)
		}
		if disableForeignKeys {
			if _, err := database.Exec(`PRAGMA foreign_keys = ON`); err != nil {
				return fmt.Errorf("restore foreign keys after migration %s: %w", version, err)
			}
		}
	}
	return nil
}

func seedCategories(database *sql.DB) error {
	transaction, err := database.Begin()
	if err != nil {
		return fmt.Errorf("begin category seed: %w", err)
	}
	for _, category := range defaultCategories {
		if _, err := transaction.Exec(`INSERT OR IGNORE INTO categories(name, created_at) VALUES(?, ?)`, category, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			transaction.Rollback()
			return fmt.Errorf("seed category %q: %w", category, err)
		}
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit category seed: %w", err)
	}
	return nil
}
