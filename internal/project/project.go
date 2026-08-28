// Package project locates the repository and Cassor state for a command.
package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const StateDirectory = ".cassor"

// FindRepositoryRoot searches from start toward the filesystem root for Git
// metadata. It supports ordinary repositories and worktrees (.git files).
func FindRepositoryRoot(start string) (string, error) {
	directory, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}

	for {
		metadataPath := filepath.Join(directory, ".git")
		if _, err := os.Stat(metadataPath); err == nil {
			return directory, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("inspect %s: %w", metadataPath, err)
		}

		parent := filepath.Dir(directory)
		if parent == directory {
			return "", fmt.Errorf("no Git repository found from %s", start)
		}
		directory = parent
	}
}

// FindStateDirectory searches upward from start for an initialized Cassor
// project, allowing commands to be run from nested repository directories.
func FindStateDirectory(start string) (string, error) {
	directory, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}

	for {
		stateDir := filepath.Join(directory, StateDirectory)
		info, err := os.Stat(stateDir)
		if err == nil && info.IsDir() {
			return stateDir, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("inspect %s: %w", stateDir, err)
		}

		parent := filepath.Dir(directory)
		if parent == directory {
			return "", fmt.Errorf("no Cassor project found from %s", start)
		}
		directory = parent
	}
}
