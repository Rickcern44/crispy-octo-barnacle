package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRepositoryRootFromNestedDirectory(t *testing.T) {
	temporary := t.TempDir()
	root := filepath.Join(temporary, "repository")
	nested := filepath.Join(root, "one", "two")
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	actual, err := FindRepositoryRoot(nested)
	if err != nil {
		t.Fatalf("FindRepositoryRoot() error = %v", err)
	}
	if actual != root {
		t.Fatalf("FindRepositoryRoot() = %q, want %q", actual, root)
	}
}

func TestFindStateDirectoryFromNestedDirectory(t *testing.T) {
	temporary := t.TempDir()
	root := filepath.Join(temporary, "repository")
	nested := filepath.Join(root, "one", "two")
	if err := os.MkdirAll(filepath.Join(root, StateDirectory), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	actual, err := FindStateDirectory(nested)
	if err != nil {
		t.Fatalf("FindStateDirectory() error = %v", err)
	}
	if actual != filepath.Join(root, StateDirectory) {
		t.Fatalf("FindStateDirectory() = %q, want %q", actual, filepath.Join(root, StateDirectory))
	}
}
