// Package config reads and writes Cassor's small project configuration file.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const FileName = "config.toml"

// Config identifies the project represented by a Cassor state directory.
type Config struct {
	Name string
}

// Write stores config at path. The compact format is deliberately limited to
// values Cassor owns, avoiding a dependency for this initial configuration.
func Write(path string, config Config) error {
	if strings.TrimSpace(config.Name) == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if strings.ContainsAny(config.Name, "\r\n") {
		return fmt.Errorf("project name cannot contain a newline")
	}

	contents := "name = " + strconv.Quote(config.Name) + "\n"
	return os.WriteFile(path, []byte(contents), 0o644)
}

// Path returns the standard configuration path within stateDir.
func Path(stateDir string) string {
	return filepath.Join(stateDir, FileName)
}

// Read loads the configuration format written by Write.
func Read(path string) (Config, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	line := strings.TrimSpace(string(contents))
	const prefix = "name = "
	if !strings.HasPrefix(line, prefix) {
		return Config{}, fmt.Errorf("configuration has no name")
	}
	name, err := strconv.Unquote(strings.TrimSpace(strings.TrimPrefix(line, prefix)))
	if err != nil || strings.TrimSpace(name) == "" {
		return Config{}, fmt.Errorf("configuration has an invalid name")
	}
	return Config{Name: name}, nil
}
