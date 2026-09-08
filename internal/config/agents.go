package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const AgentsFileName = "agents.toml"

// AgentsPath is the project-local runtime capability configuration.
func AgentsPath(stateDir string) string { return filepath.Join(stateDir, AgentsFileName) }

// RoleModels contains one model or capability class for every Cassor role.
type RoleModels struct {
	Orchestrator   string
	Discovery      string
	Implementation string
	Verification   string
}

// AgentsConfig is the project-local worker configuration. Models remain
// portable capability classes, while Codex maps them to model identifiers the
// current runtime may select for a worker.
type AgentsConfig struct {
	Models RoleModels
	Codex  RoleModels
}

var roleNames = map[string]func(*RoleModels, string){
	"orchestrator":   func(models *RoleModels, value string) { models.Orchestrator = value },
	"discovery":      func(models *RoleModels, value string) { models.Discovery = value },
	"implementation": func(models *RoleModels, value string) { models.Implementation = value },
	"verification":   func(models *RoleModels, value string) { models.Verification = value },
}

// WriteDefaultAgents stores portable capability classes and the default Codex
// routing. Runtime adapters may add scoped overrides without changing these
// defaults.
func WriteDefaultAgents(path string) error {
	const contents = "# Cassor role classes and runtime-specific worker routing.\n# Codex uses this mapping only when explicit worker-model routing is available.\n[models]\norchestrator = \"frontier\"\ndiscovery = \"frontier\"\nimplementation = \"economical-coding\"\nverification = \"economical-coding\"\n\n[runtimes.codex]\norchestrator = \"gpt-5.6-sol\"\ndiscovery = \"gpt-5.6-sol\"\nimplementation = \"gpt-5.6-luna\"\nverification = \"gpt-5.6-luna\"\n"
	return os.WriteFile(path, []byte(contents), 0o644)
}

// ReadAgents parses Cassor's intentionally small agents.toml surface.
func ReadAgents(path string) (AgentsConfig, error) {
	contents, err := os.Open(path)
	if err != nil {
		return AgentsConfig{}, err
	}
	defer contents.Close()

	var result AgentsConfig
	section := ""
	scanner := bufio.NewScanner(contents)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}
		key, rawValue, found := strings.Cut(line, "=")
		if !found {
			return AgentsConfig{}, fmt.Errorf("agents configuration line %d must be key = value", lineNumber)
		}
		set, known := roleNames[strings.TrimSpace(key)]
		if !known || (section != "models" && section != "runtimes.codex") {
			continue
		}
		value, err := strconv.Unquote(strings.TrimSpace(rawValue))
		if err != nil {
			return AgentsConfig{}, fmt.Errorf("agents configuration line %d has an invalid string: %w", lineNumber, err)
		}
		if section == "models" {
			set(&result.Models, value)
		} else {
			set(&result.Codex, value)
		}
	}
	if err := scanner.Err(); err != nil {
		return AgentsConfig{}, fmt.Errorf("read agents configuration: %w", err)
	}
	return result, nil
}

// ValidateCodexRouting checks that all portable and Codex role assignments
// needed by the installed Cassor skill are present.
func ValidateCodexRouting(config AgentsConfig) error {
	for _, group := range []struct {
		name   string
		models RoleModels
	}{
		{"[models]", config.Models},
		{"[runtimes.codex]", config.Codex},
	} {
		for _, role := range []struct {
			name  string
			value string
		}{
			{"orchestrator", group.models.Orchestrator},
			{"discovery", group.models.Discovery},
			{"implementation", group.models.Implementation},
			{"verification", group.models.Verification},
		} {
			if strings.TrimSpace(role.value) == "" {
				return fmt.Errorf("agents configuration is missing %s.%s", group.name, role.name)
			}
		}
	}
	return nil
}
