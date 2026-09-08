package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rickcern44/cassor/internal/config"
	"github.com/rickcern44/cassor/internal/project"
	"github.com/rickcern44/cassor/internal/store"
)

func TestItemDossierAndRelationshipCommands(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
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
	database, err := store.Open(filepath.Join(state, store.DatabaseFileName))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	capability, err := store.AddItem(database, "Authentication", "", "Needed", "Now", "")
	if err != nil {
		t.Fatal(err)
	}
	change, err := store.AddItem(database, "Passkeys", "", "Needed", "Next", "")
	if err != nil {
		t.Fatal(err)
	}

	command := NewRootCommand()
	command.SetArgs([]string{"item", "update", "1", "--feature-type", "Capability"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	updated, err := store.GetItem(database, capability.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.FeatureType != "Capability" || updated.CurrentState != "" {
		t.Fatalf("updated feature = %#v", updated)
	}

	command = NewRootCommand()
	command.SetArgs([]string{"item", "relationship", "add", "--from", "2", "--to", "1", "--type", "extends"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	relationships, err := store.ListFeatureRelationships(database, change.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(relationships) != 1 || relationships[0].TargetItemID != capability.ID {
		t.Fatalf("relationships = %#v", relationships)
	}
}
