package check

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rickcern44/cassor/internal/config"
	"github.com/rickcern44/cassor/internal/project"
	"github.com/rickcern44/cassor/internal/store"
)

func TestRunAcceptsValidStateAndRejectsBrokenLifecycle(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, project.StateDirectory)
	if err := os.Mkdir(state, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := config.Write(config.Path(state), config.Config{Name: "Check"}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(state, store.DatabaseFileName)
	if err := store.Initialize(path); err != nil {
		t.Fatal(err)
	}
	if err := Run(state); err != nil {
		t.Fatalf("Run() valid state error = %v", err)
	}
	database, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	item, err := store.AddItem(database, "Item", "", "Needed", "Now", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SetItemStatus(database, item.ID, "Ready"); err != nil {
		t.Fatal(err)
	}
	plan, err := store.CreatePlan(database, item.ID, "Plan")
	if err != nil {
		t.Fatal(err)
	}
	task, err := store.AddTask(database, plan.ID, "Task", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`UPDATE tasks SET status='Done' WHERE id=?`, task.ID); err != nil {
		t.Fatal(err)
	}
	if err := Run(state); err == nil {
		t.Fatal("Run() accepted a completed task without lifecycle details")
	}
}

func TestRunRejectsInvalidDossierRecords(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, project.StateDirectory)
	if err := os.Mkdir(state, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := config.Write(config.Path(state), config.Config{Name: "Check"}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(state, store.DatabaseFileName)
	if err := store.Initialize(path); err != nil {
		t.Fatal(err)
	}
	database, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	change, err := store.AddItem(database, "Change", "", "Needed", "Now", "")
	if err != nil {
		t.Fatal(err)
	}
	otherChange, err := store.AddItem(database, "Other change", "", "Needed", "Now", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO feature_change_links(change_item_id,capability_item_id,created_at) VALUES(?,?,?)`, change.ID, otherChange.ID, "test"); err != nil {
		t.Fatal(err)
	}
	if err := Run(state); err == nil {
		t.Fatal("check accepted a change link whose target is not a capability")
	}
}

func TestRunAcceptsCompletedItemWithApprovedPlan(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, project.StateDirectory)
	if err := os.Mkdir(state, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := config.Write(config.Path(state), config.Config{Name: "Check"}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(state, store.DatabaseFileName)
	if err := store.Initialize(path); err != nil {
		t.Fatal(err)
	}
	database, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	item, err := store.AddItem(database, "Item", "", "Needed", "Now", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.TransitionItemStatus(database, item.ID, "Ready"); err != nil {
		t.Fatal(err)
	}
	plan, err := store.CreatePlan(database, item.ID, "Plan")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.AddAcceptanceCriterionForPlan(database, plan.ID, "done", "The item is complete", "", true); err != nil {
		t.Fatal(err)
	}
	task, err := store.AddTask(database, plan.ID, "Task", "", []string{"go test ./..."})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.ApprovePlan(database, plan.ID); err != nil {
		t.Fatal(err)
	}
	if err := store.SetTaskStatus(database, task.ID, "In Progress", ""); err != nil {
		t.Fatal(err)
	}
	if err := store.SetTaskStatus(database, task.ID, "Done", "verified"); err != nil {
		t.Fatal(err)
	}
	criteria, err := store.ListAcceptanceCriteria(database)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.VerifyAcceptanceCriterion(database, criteria[0].ID, "Passed", "go test", "verified"); err != nil {
		t.Fatal(err)
	}
	if err := store.TransitionItemStatus(database, item.ID, "In Progress"); err != nil {
		t.Fatal(err)
	}
	if err := store.TransitionItemStatus(database, item.ID, "Done"); err != nil {
		t.Fatal(err)
	}
	if err := Run(state); err != nil {
		t.Fatalf("Run() completed item error = %v", err)
	}
}
