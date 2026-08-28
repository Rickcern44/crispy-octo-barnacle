package config

import (
	"os"
	"path/filepath"
)

const AgentsFileName = "agents.toml"

// AgentsPath is the project-local runtime capability configuration.
func AgentsPath(stateDir string) string { return filepath.Join(stateDir, AgentsFileName) }

// WriteDefaultAgents stores portable capability classes. Runtime adapters may
// add scoped overrides without changing these defaults.
func WriteDefaultAgents(path string) error {
	const contents = "[models]\norchestrator = \"frontier\"\ndiscovery = \"economical\"\nimplementation = \"economical-coding\"\nverification = \"economical-coding\"\n"
	return os.WriteFile(path, []byte(contents), 0o644)
}
