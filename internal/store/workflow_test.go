package store

import (
	"path/filepath"
	"testing"
)

func TestWorkflowEnforcesApprovalAndImmutablePlans(t *testing.T) {
	path := filepath.Join(t.TempDir(), DatabaseFileName)
	if err := Initialize(path); err != nil {
		t.Fatal(err)
	}
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	item, err := AddItem(database, "Ship init", "", "Needed", "Now", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CreatePlan(database, item.ID, "implementation"); err == nil {
		t.Fatal("CreatePlan() succeeded for proposed item")
	}
	if err := SetItemStatus(database, item.ID, "Ready"); err != nil {
		t.Fatal(err)
	}
	plan, err := CreatePlan(database, item.ID, "implementation")
	if err != nil {
		t.Fatal(err)
	}
	task, err := AddTask(database, plan.ID, "Implement", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := SetTaskStatus(database, task.ID, "In Progress", ""); err == nil {
		t.Fatal("SetTaskStatus() started task before plan approval")
	}
	if err := ApprovePlan(database, plan.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := AddTask(database, plan.ID, "Late work", ""); err == nil {
		t.Fatal("AddTask() modified approved plan")
	}
	if err := SetTaskStatus(database, task.ID, "In Progress", ""); err != nil {
		t.Fatal(err)
	}
	if err := SetTaskStatus(database, task.ID, "Done", "done"); err != nil {
		t.Fatal(err)
	}
}
