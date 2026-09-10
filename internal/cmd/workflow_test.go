package cmd

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rickcern44/cassor/internal/config"
	"github.com/rickcern44/cassor/internal/project"
	"github.com/rickcern44/cassor/internal/store"
)

func TestItemNextCommandReportsNoActionableWork(t *testing.T) {
	root, database := workflowTestProject(t)
	defer database.Close()
	_ = root

	command := NewRootCommand()
	var outputBuffer bytes.Buffer
	command.SetOut(&outputBuffer)
	command.SetArgs([]string{"item", "next"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(outputBuffer.String()); got != "No actionable work found." {
		t.Fatalf("output = %q", got)
	}
}

func TestItemNextCommandReportsScopedContext(t *testing.T) {
	root, database := workflowTestProject(t)
	defer database.Close()
	_ = root
	packet := store.PlanPacket{
		RoadmapItem: store.PacketRoadmapItem{Title: "Implement next", Category: "Needed", Horizon: "Now"},
		Goal:        "Implement the next task",
		AcceptanceCriteria: []store.PacketCriterion{{
			Title: "The task is verified",
		}},
		Tasks: []store.PacketTask{{Title: "Next task", Verification: []string{"go test ./..."}}},
	}
	if _, err := store.RecordApprovedPlan(database, packet, "approved"); err != nil {
		t.Fatal(err)
	}

	command := NewRootCommand()
	var outputBuffer bytes.Buffer
	command.SetOut(&outputBuffer)
	command.SetArgs([]string{"item", "next"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	output := outputBuffer.String()
	if !strings.Contains(output, "Next task") || !strings.Contains(output, "cassor context --item 1 --role implementation --max-bytes 12000") {
		t.Fatalf("output = %q", output)
	}
}

func workflowTestProject(t *testing.T) (string, *sql.DB) {
	t.Helper()
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
	return root, database
}

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
	command.SetArgs([]string{"item", "update", "1", "--current-state", "deprecated"})
	if err := command.Execute(); err == nil {
		t.Fatal("item update accepted the deprecated --current-state flag")
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
