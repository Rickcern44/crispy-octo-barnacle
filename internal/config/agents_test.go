package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteDefaultAgentsProvidesCodexRouting(t *testing.T) {
	path := filepath.Join(t.TempDir(), AgentsFileName)
	if err := WriteDefaultAgents(path); err != nil {
		t.Fatal(err)
	}
	configuration, err := ReadAgents(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateCodexRouting(configuration); err != nil {
		t.Fatal(err)
	}
	if configuration.Codex.Orchestrator != "gpt-5.6-sol" || configuration.Codex.Discovery != "gpt-5.6-sol" {
		t.Fatalf("planning routing = %+v", configuration.Codex)
	}
	if configuration.Codex.Implementation != "gpt-5.6-luna" || configuration.Codex.Verification != "gpt-5.6-luna" {
		t.Fatalf("worker routing = %+v", configuration.Codex)
	}
}

func TestValidateCodexRoutingReportsMissingRole(t *testing.T) {
	configuration := AgentsConfig{Models: RoleModels{"frontier", "frontier", "economical-coding", "economical-coding"}, Codex: RoleModels{Orchestrator: "gpt-5.6-sol"}}
	err := ValidateCodexRouting(configuration)
	if err == nil || !strings.Contains(err.Error(), "[runtimes.codex].discovery") {
		t.Fatalf("validation error = %v", err)
	}
}
