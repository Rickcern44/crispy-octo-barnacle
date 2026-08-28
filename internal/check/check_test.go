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
	if err := store.SetItemStatus(database, item.ID, "Approved"); err != nil {
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
	if _, err := database.Exec(`UPDATE tasks SET status='Completed' WHERE id=?`, task.ID); err != nil {
		t.Fatal(err)
	}
	if err := Run(state); err == nil {
		t.Fatal("Run() accepted a completed task without lifecycle details")
	}
}
