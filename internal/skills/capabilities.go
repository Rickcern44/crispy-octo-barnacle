package skills

import (
	"os"
	"path/filepath"

	"github.com/rickcern44/cassor/internal/config"
)

// RuntimeCapability is the observable adapter boundary for one supported
// runtime. Status values are deliberately explicit so unavailable delegation
// is not mistaken for a failed installation or an available worker.
type RuntimeCapability struct {
	Agent        string            `json:"agent"`
	Adapter      string            `json:"adapter"`
	Installation string            `json:"installation"`
	Capabilities map[string]string `json:"capabilities"`
	Fallback     string            `json:"fallback"`
}

// CapabilityReport is a read-only snapshot of the project adapter surfaces.
type CapabilityReport struct {
	Runtimes []RuntimeCapability `json:"runtimes"`
}

// Capabilities reports only capabilities Cassor can observe from project
// configuration and managed skill assets. It never probes or infers hidden
// runtime worker support.
func Capabilities(repositoryRoot, stateDir string) (CapabilityReport, error) {
	manifest, err := readManifest(ManifestPath(stateDir))
	if err != nil {
		return CapabilityReport{}, err
	}
	agents, agentsErr := config.ReadAgents(config.AgentsPath(stateDir))
	values := make([]RuntimeCapability, 0, 3)
	for _, agent := range []string{"codex", "claude-code", "copilot"} {
		assets, err := assetsForAgent(agent)
		if err != nil {
			return CapabilityReport{}, err
		}
		installation := findInstallation(manifest, agent, "project")
		capability := RuntimeCapability{
			Agent:        agent,
			Adapter:      "supported",
			Installation: installationState(repositoryRoot, installation, assets),
			Capabilities: map[string]string{
				"worker_isolation": "unavailable",
				"model_routing":    "unavailable",
			},
			Fallback: "single-agent orchestrator with Cassor CLI authority",
		}
		if agent == "codex" && agentsErr == nil && config.ValidateCodexRouting(agents) == nil {
			capability.Capabilities["model_routing"] = "configured"
		}
		values = append(values, capability)
	}
	return CapabilityReport{Runtimes: values}, nil
}

func installationState(repositoryRoot string, installation *Installation, assets map[string][]byte) string {
	if installation == nil {
		return "missing"
	}
	destination := filepath.Join(repositoryRoot, filepath.FromSlash(installation.Destination))
	managed := make(map[string]string, len(installation.Files))
	for _, file := range installation.Files {
		managed[file.Path] = file.SHA256
		contents, err := os.ReadFile(filepath.Join(destination, file.Path))
		if err != nil || hash(contents) != file.SHA256 {
			return "drifted"
		}
	}
	for path, contents := range assets {
		if managed[path] == "" || managed[path] != hash(contents) {
			return "update-available"
		}
	}
	return "current"
}
