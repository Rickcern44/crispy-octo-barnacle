package skills

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rickcern44/cassor/internal/config"
)

func TestVerifyManagedFileExplainsAvailableUpdate(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "SKILL.md"), []byte("installed"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := verifyManagedFile(root, Installation{Files: []ManagedFile{{Path: "SKILL.md", SHA256: hash([]byte("installed"))}}}, "SKILL.md", []byte("new version"))
	if err == nil || !strings.Contains(err.Error(), "Cassor skill update available") || !strings.Contains(err.Error(), "cassor skills update") {
		t.Fatalf("update error = %v", err)
	}
}

func TestInstallIsIdempotentAndTracksManifest(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, ".cassor")
	if err := os.Mkdir(state, 0o755); err != nil {
		t.Fatal(err)
	}
	result, err := Install(root, state, InstallOptions{Agent: "codex", Scope: "project"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Changes) == 0 {
		t.Fatal("Install() reported no created files")
	}
	if _, err := os.Stat(filepath.Join(root, ".agents", "skills", "cassor", "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	second, err := Install(root, state, InstallOptions{Agent: "codex", Scope: "project"})
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Changes) != 0 {
		t.Fatalf("repeated install changes = %+v", second.Changes)
	}
	manifest, err := Status(root, state)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Installations) != 1 {
		t.Fatalf("installations = %+v", manifest.Installations)
	}
}

func TestInstallProtectsConflictingDestinationAndDryRunDoesNotWrite(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, ".cassor")
	if err := os.Mkdir(state, 0o755); err != nil {
		t.Fatal(err)
	}
	dryRun, err := Install(root, state, InstallOptions{Agent: "codex", Scope: "project", DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(dryRun.Changes) == 0 {
		t.Fatal("dry run has no changes")
	}
	if _, err := os.Stat(dryRun.Destination); !os.IsNotExist(err) {
		t.Fatalf("dry run created destination: %v", err)
	}
	if err := os.MkdirAll(dryRun.Destination, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(root, state, InstallOptions{Agent: "codex", Scope: "project"}); err == nil {
		t.Fatal("Install() accepted conflicting destination")
	}
}

func TestUpdateAndUninstallPreserveModifiedFiles(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, ".cassor")
	if err := os.Mkdir(state, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(root, state, InstallOptions{Agent: "codex", Scope: "project"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Update(root, state, "codex", "project"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ".agents", "skills", "cassor", "SKILL.md")
	if err := os.WriteFile(path, []byte("custom"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Uninstall(root, state, "codex", "project")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Warnings) == 0 {
		t.Fatal("Uninstall() did not preserve modified file")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func TestInstallAllPreflightsConflicts(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, ".cassor")
	if err := os.Mkdir(state, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".github", "skills", "cassor"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(root, state, InstallOptions{Agent: "all", Scope: "project"}); err == nil {
		t.Fatal("Install(all) accepted a conflict")
	}
	if _, err := os.Stat(filepath.Join(root, ".agents", "skills", "cassor")); !os.IsNotExist(err) {
		t.Fatalf("all install wrote before conflict: %v", err)
	}
}

func TestInstallAllCreatesNativeDestinations(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, ".cassor")
	if err := os.Mkdir(state, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(root, state, InstallOptions{Agent: "all", Scope: "project"}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{".agents/skills/cassor/SKILL.md", ".claude/skills/cassor/SKILL.md", ".github/skills/cassor/SKILL.md"} {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".claude/skills/cassor/references/adapters/claude-code.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".github/skills/cassor/references/adapters/copilot.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".agents/skills/cassor/references/adapters/claude-code.md")); !os.IsNotExist(err) {
		t.Fatalf("Codex received a foreign adapter: %v", err)
	}
	manifest, err := Status(root, state)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Installations) != 3 {
		t.Fatalf("installations=%d", len(manifest.Installations))
	}
}

func TestRuntimeAssetsKeepPortableCoreAndAgentSpecificAdapters(t *testing.T) {
	core, err := portableAssets()
	if err != nil {
		t.Fatal(err)
	}
	claude, err := assetsForAgent("claude-code")
	if err != nil {
		t.Fatal(err)
	}
	copilot, err := assetsForAgent("copilot")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := core["references/adapters/claude-code.md"]; ok {
		t.Fatal("portable core unexpectedly contains Claude adapter")
	}
	if string(claude["SKILL.md"]) != string(copilot["SKILL.md"]) {
		t.Fatal("runtime adapters diverged in the portable skill")
	}
	if _, ok := claude["references/adapters/claude-code.md"]; !ok {
		t.Fatal("Claude adapter asset missing")
	}
	if _, ok := claude["references/adapters/copilot.md"]; ok {
		t.Fatal("Claude assets contain Copilot adapter")
	}
	if _, ok := copilot["references/adapters/copilot.md"]; !ok {
		t.Fatal("Copilot adapter asset missing")
	}
}

func TestClaudeAdapterRemainsProjectScoped(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, ".cassor")
	if err := os.Mkdir(state, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(root, state, InstallOptions{Agent: "claude-code", Scope: "project"}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		".claude/skills/cassor/references/adapters/claude-code.md",
		".claude/skills/cassor/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range []string{".claude/agents", ".claude/settings.json", ".claude/hooks"} {
		if _, err := os.Stat(filepath.Join(root, path)); !os.IsNotExist(err) {
			t.Fatalf("Claude global or optional configuration was created at %s: %v", path, err)
		}
	}
	adapter, err := os.ReadFile(filepath.Join(root, ".claude/skills/cassor/references/adapters/claude-code.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(adapter), "one orchestrator") || !strings.Contains(string(adapter), "global Claude configuration") {
		t.Fatalf("Claude adapter lacks safe fallback guidance: %s", adapter)
	}
}

func TestCopilotAdapterRemainsProjectScoped(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, ".cassor")
	if err := os.Mkdir(state, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(root, state, InstallOptions{Agent: "copilot", Scope: "project"}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		".github/skills/cassor/references/adapters/copilot.md",
		".github/skills/cassor/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range []string{".github/copilot-instructions.md", ".github/agents", ".github/hooks"} {
		if _, err := os.Stat(filepath.Join(root, path)); !os.IsNotExist(err) {
			t.Fatalf("Copilot global or optional configuration was created at %s: %v", path, err)
		}
	}
	adapter, err := os.ReadFile(filepath.Join(root, ".github/skills/cassor/references/adapters/copilot.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(adapter), "one orchestrator") || !strings.Contains(string(adapter), "repository-wide Copilot instructions") {
		t.Fatalf("Copilot adapter lacks safe fallback guidance: %s", adapter)
	}
}

func TestCapabilitiesReportInstallationAndSafeFallback(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, ".cassor")
	if err := os.Mkdir(state, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(root, state, InstallOptions{Agent: "claude-code", Scope: "project"}); err != nil {
		t.Fatal(err)
	}
	report, err := Capabilities(root, state)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Runtimes) != 3 {
		t.Fatalf("runtime count = %d", len(report.Runtimes))
	}
	byAgent := map[string]RuntimeCapability{}
	for _, runtime := range report.Runtimes {
		byAgent[runtime.Agent] = runtime
	}
	if byAgent["claude-code"].Installation != "current" {
		t.Fatalf("Claude installation = %+v", byAgent["claude-code"])
	}
	if byAgent["copilot"].Installation != "missing" || byAgent["codex"].Installation != "missing" {
		t.Fatalf("missing installations = %+v", byAgent)
	}
	for _, runtime := range report.Runtimes {
		if runtime.Capabilities["worker_isolation"] != "unavailable" || runtime.Fallback == "" {
			t.Fatalf("unsafe fallback report = %+v", runtime)
		}
	}
}

func TestCodexInstallDeliversAdaptiveDispatchPolicyFromEmbeddedAssets(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, ".cassor")
	if err := os.Mkdir(state, 0o755); err != nil {
		t.Fatal(err)
	}
	assets, err := portableAssets()
	if err != nil {
		t.Fatal(err)
	}
	for path, required := range map[string][]string{
		"SKILL.md": {
			"Follow the proportionate SDD-lite lifecycle",
			"explicit approval before mutation",
			"references/adaptive-orchestration.md",
			"cassor plan record --file",
		},
		"references/adaptive-orchestration.md": {
			"subagent creation",
			"Parallel read-only discovery workers",
			"Exactly one implementation worker",
			".cassor/agents.toml",
		},
	} {
		contents, ok := assets[path]
		if !ok {
			t.Fatalf("embedded assets missing %s", path)
		}
		for _, text := range required {
			if !strings.Contains(string(contents), text) {
				t.Errorf("embedded %s missing adaptive dispatch policy %q", path, text)
			}
		}
	}

	if _, err := Install(root, state, InstallOptions{Agent: "codex", Scope: "project"}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"SKILL.md", "references/adaptive-orchestration.md", "references/protocol.md"} {
		installed, err := os.ReadFile(filepath.Join(root, ".agents", "skills", "cassor", path))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(installed, assets[path]) {
			t.Errorf("installed %s differs from its embedded portable asset", path)
		}
	}
}

func TestDoctorWarnsWhenCodexRoutingIsMissing(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, ".cassor")
	if err := os.Mkdir(state, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config.AgentsPath(state), []byte("[models]\norchestrator = \"frontier\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := Doctor(root, state)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != "warning" || len(report.Warnings) == 0 || !strings.Contains(strings.Join(report.Warnings, "\n"), "Codex model routing is incomplete") {
		t.Fatalf("doctor report = %+v", report)
	}
}

func TestPortableSkillKeepsRuntimeDetailsOnDemand(t *testing.T) {
	assets, err := portableAssets()
	if err != nil {
		t.Fatal(err)
	}
	skill := string(assets["SKILL.md"])
	for _, required := range []string{"name: cassor", "cassor context --json", "cassor check", "references/adaptive-orchestration.md", "cassor plan record", "cassor run-report"} {
		if !strings.Contains(skill, required) {
			t.Errorf("SKILL.md missing %q", required)
		}
	}
	for _, forbidden := range []string{"gpt-", ".cassor/agents.toml", "subagent creation", "Parallel read-only discovery workers"} {
		if strings.Contains(skill, forbidden) {
			t.Errorf("SKILL.md eagerly contains runtime detail %q", forbidden)
		}
	}
	for _, path := range []string{"references/protocol.md", "references/schemas.md", "references/adaptive-orchestration.md", "references/roles/discovery.md", "references/roles/implementation.md", "references/roles/verification.md"} {
		if _, ok := assets[path]; !ok {
			t.Errorf("required skill reference missing: %s", path)
		}
	}
	if !strings.Contains(string(assets["references/adaptive-orchestration.md"]), ".cassor/agents.toml") {
		t.Fatal("adaptive reference missing project capability configuration")
	}
}
