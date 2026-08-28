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
