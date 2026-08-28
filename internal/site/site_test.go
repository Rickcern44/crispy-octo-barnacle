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

func TestBuildIsDeterministic(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, project.StateDirectory)
	if err := os.Mkdir(state, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := config.Write(config.Path(state), config.Config{Name: "Demo"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Initialize(filepath.Join(state, store.DatabaseFileName)); err != nil {
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
	if !strings.Contains(string(initial), "const data=") {
		t.Fatal("roadmap does not embed data")
	}
}
