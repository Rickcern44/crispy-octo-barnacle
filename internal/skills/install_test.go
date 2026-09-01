package skills

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	manifest, err := Status(root, state)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Installations) != 3 {
		t.Fatalf("installations=%d", len(manifest.Installations))
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
			"subagents, parallel execution, isolated worker contexts, compact structured returns, and explicit worker-model routing",
			"Parallel workers are read-only.",
			"Exactly one implementation worker; never concurrent implementation workers.",
			"Perform the same role contracts sequentially in the orchestrator context.",
		},
		"references/protocol.md": {
			"Treat an unknown capability as unavailable.",
			"Parallel work requires available parallel execution and isolated contexts.",
			"Never run more than one implementation worker",
			"The orchestrator alone retains those authorities.",
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
	for _, path := range []string{"SKILL.md", "references/protocol.md"} {
		installed, err := os.ReadFile(filepath.Join(root, ".agents", "skills", "cassor", path))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(installed, assets[path]) {
			t.Errorf("installed %s differs from its embedded portable asset", path)
		}
	}
}
