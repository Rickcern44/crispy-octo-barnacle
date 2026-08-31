package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rickcern44/cassor/internal/config"
	"github.com/rickcern44/cassor/internal/project"
	"github.com/rickcern44/cassor/internal/store"
)

func TestBuildGeneratesDeterministicFeatureRoutes(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, project.StateDirectory)
	if err := os.Mkdir(state, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"CASSOR_CODEX_HANDOFF.md", "CASSOR_PLAN_PACKET_SCHEMA.md", "CASSOR_SKILLS_SPEC.md"} {
		if err := os.WriteFile(filepath.Join(root, "docs", name), []byte("# "+name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := config.Write(config.Path(state), config.Config{Name: "Demo"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Initialize(filepath.Join(state, store.DatabaseFileName)); err != nil {
		t.Fatal(err)
	}
	database, err := store.Open(filepath.Join(state, store.DatabaseFileName))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if _, err := store.AddItem(database, "Document static routes", "Publish feature detail pages.", "Needed", "Now", ""); err != nil {
		t.Fatal(err)
	}
	first, err := Build(root, state)
	if err != nil {
		t.Fatal(err)
	}
	initial, err := os.ReadFile(first)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Build(root, state)
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := os.ReadFile(second)
	if err != nil {
		t.Fatal(err)
	}
	if string(initial) != string(repeated) {
		t.Fatal("Build() output changed without state changes")
	}
	if !strings.Contains(string(initial), "/roadmap/rm-1/") {
		t.Fatal("roadmap does not link to generated feature detail route")
	}
	feature, err := os.ReadFile(filepath.Join(root, sourceDocsDirectory, "roadmap", "rm-1.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(feature), "Document static routes") {
		t.Fatal("feature detail source does not include item title")
	}
}
