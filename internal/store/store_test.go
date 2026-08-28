package store

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestInitializeAppliesMigrationsAndSeedsCategories(t *testing.T) {
	path := filepath.Join(t.TempDir(), DatabaseFileName)
	if err := Initialize(path); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	database, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM categories`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != len(defaultCategories) {
		t.Fatalf("category count = %d, want %d", count, len(defaultCategories))
	}
}
