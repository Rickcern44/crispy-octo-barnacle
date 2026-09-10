package store

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestSelectNextTaskReturnsNoWork(t *testing.T) {
	database := testNextDatabase(t)
	selection, err := SelectNextTask(database)
	if err != nil {
		t.Fatal(err)
	}
	if selection != nil {
		t.Fatalf("selection = %#v, want nil", selection)
	}
}

func TestSelectNextTaskReturnsOldestReadyWork(t *testing.T) {
	database := testNextDatabase(t)
	first := recordNextPlan(t, database, "First ready item", "First task")
	second := recordNextPlan(t, database, "Second ready item", "Second task")

	selection, err := SelectNextTask(database)
	if err != nil {
		t.Fatal(err)
	}
	if selection == nil || selection.ItemID != first.Plan.ItemID || selection.TaskID != first.Tasks[0].ID {
		t.Fatalf("selection = %#v, want first item/task (%d/%d)", selection, first.Plan.ItemID, first.Tasks[0].ID)
	}
	if selection.ItemID == second.Plan.ItemID || selection.ScopedContextCommand != "cassor context --item 1 --role implementation --max-bytes 12000" {
		t.Fatalf("selection = %#v", selection)
	}
}

func TestSelectNextTaskPrioritizesInProgressWork(t *testing.T) {
	database := testNextDatabase(t)
	ready := recordNextPlan(t, database, "Ready item", "Ready task")
	inProgress := recordNextPlan(t, database, "In progress item", "Active task")
	if err := SetItemStatus(database, inProgress.Plan.ItemID, "In Progress"); err != nil {
		t.Fatal(err)
	}
	if err := SetTaskStatus(database, inProgress.Tasks[0].ID, "In Progress", ""); err != nil {
		t.Fatal(err)
	}

	selection, err := SelectNextTask(database)
	if err != nil {
		t.Fatal(err)
	}
	if selection == nil || selection.ItemID != inProgress.Plan.ItemID || selection.TaskID != inProgress.Tasks[0].ID || selection.TaskStatus != "In Progress" {
		t.Fatalf("selection = %#v, ready=%#v", selection, ready)
	}
}

func testNextDatabase(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), DatabaseFileName)
	if err := Initialize(path); err != nil {
		t.Fatal(err)
	}
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

func recordNextPlan(t *testing.T, database *sql.DB, itemTitle, taskTitle string) RecordedPlan {
	t.Helper()
	packet := PlanPacket{
		RoadmapItem:        PacketRoadmapItem{Title: itemTitle, Category: "Needed", Horizon: "Now"},
		Goal:               "Implement the selected task",
		AcceptanceCriteria: []PacketCriterion{{Title: "The task is verified"}},
		Tasks:              []PacketTask{{Title: taskTitle, Verification: []string{"go test ./..."}}},
	}
	recorded, err := RecordApprovedPlan(database, packet, "approved")
	if err != nil {
		t.Fatal(err)
	}
	return recorded
}
