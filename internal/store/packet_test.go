package store

import (
	"path/filepath"
	"testing"
)

func TestRecordApprovedPlanIsAtomicAndPreservesApprovalEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), DatabaseFileName)
	if err := Initialize(path); err != nil {
		t.Fatal(err)
	}
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	packet := PlanPacket{RoadmapItem: PacketRoadmapItem{Title: "Skills", Category: "Needed", Horizon: "Now"}, Goal: "Install the skills layer", AcceptanceCriteria: []PacketCriterion{{ID: "contract", Title: "The skill is installable"}}, Tasks: []PacketTask{{Title: "Add protocol", Verification: []string{"go test ./..."}}}}
	recorded, err := RecordApprovedPlan(database, packet, "Approved in Codex")
	if err != nil {
		t.Fatal(err)
	}
	if recorded.Plan.Status != "Approved" || recorded.Plan.ApprovalNote != "Approved in Codex" {
		t.Fatalf("recorded plan = %+v", recorded.Plan)
	}
	if len(recorded.Tasks) != 1 || recorded.Tasks[0].Status != "To Do" {
		t.Fatalf("recorded tasks = %+v", recorded.Tasks)
	}
	var items, plans, tasks int
	if err := database.QueryRow(`SELECT COUNT(*) FROM roadmap_items`).Scan(&items); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow(`SELECT COUNT(*) FROM plan_revisions`).Scan(&plans); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&tasks); err != nil {
		t.Fatal(err)
	}
	if items != 1 || plans != 1 || tasks != 1 {
		t.Fatalf("counts = %d, %d, %d", items, plans, tasks)
	}
}

func TestRecordApprovedPlanRejectsOpenQuestionsWithoutWriting(t *testing.T) {
	path := filepath.Join(t.TempDir(), DatabaseFileName)
	if err := Initialize(path); err != nil {
		t.Fatal(err)
	}
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	_, err = RecordApprovedPlan(database, PlanPacket{RoadmapItem: PacketRoadmapItem{Title: "Skills", Category: "Needed", Horizon: "Now"}, Goal: "Goal", OpenQuestions: []string{"Which runtime?"}, Tasks: []PacketTask{{Title: "Task"}}}, "")
	if err == nil {
		t.Fatal("RecordApprovedPlan() accepted open questions")
	}
	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM roadmap_items`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("roadmap item count = %d, want 0", count)
	}
}
