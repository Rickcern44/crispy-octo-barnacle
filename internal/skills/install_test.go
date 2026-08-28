package skills

import (
	"os"
	"path/filepath"
	"testing"
)

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
