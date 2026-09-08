package store

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/rickcern44/cassor/internal/migrations"
)

func openTestDatabase(t *testing.T) *sql.DB {
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

func readyItem(t *testing.T, database *sql.DB, title string) Item {
	t.Helper()
	item, err := AddItem(database, title, "", "Needed", "Now", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := SetItemStatus(database, item.ID, "Ready"); err != nil {
		t.Fatal(err)
	}
	return item
}

func TestRecordApprovedPlanRollsBackAfterTaskFailure(t *testing.T) {
	database := openTestDatabase(t)
	if _, err := database.Exec(`CREATE TRIGGER reject_contract_task BEFORE INSERT ON tasks WHEN NEW.title='reject' BEGIN SELECT RAISE(ABORT, 'test failure'); END`); err != nil {
		t.Fatal(err)
	}
	_, err := RecordApprovedPlan(database, PlanPacket{
		RoadmapItem:        PacketRoadmapItem{Title: "Atomic", Category: "Needed", Horizon: "Now"},
		Goal:               "Test atomicity",
		AcceptanceCriteria: []PacketCriterion{{ID: "contract", Title: "The contract is present"}},
		Tasks:              []PacketTask{{Title: "reject", Verification: []string{"go test ./..."}}},
	}, "")
	if err == nil {
		t.Fatal("RecordApprovedPlan() accepted the trigger failure")
	}
	for _, table := range []string{"roadmap_items", "plan_revisions", "acceptance_criteria", "tasks"} {
		var count int
		if err := database.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("%s count = %d after rollback", table, count)
		}
	}
}

func TestAmendmentActivatesOnlyNewestRevisionAndCarriesForwardWork(t *testing.T) {
	database := openTestDatabase(t)
	item := readyItem(t, database, "Amendment")
	first, err := CreatePlan(database, item.ID, "first")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AddAcceptanceCriterionForPlan(database, first.ID, "contract", "The change is verified", "", true); err != nil {
		t.Fatal(err)
	}
	oldTask, err := AddTask(database, first.ID, "Carry this", "", []string{"go test ./..."})
	if err != nil {
		t.Fatal(err)
	}
	if err := ApprovePlan(database, first.ID); err != nil {
		t.Fatal(err)
	}
	second, err := RevisePlan(database, first.ID, "amended")
	if err != nil {
		t.Fatal(err)
	}
	if err := ApprovePlan(database, second.ID); err != nil {
		t.Fatal(err)
	}
	first, _ = GetPlan(database, first.ID)
	second, _ = GetPlan(database, second.ID)
	if first.Active || !second.Active {
		t.Fatalf("active authority = first %v, second %v", first.Active, second.Active)
	}
	if err := SetTaskStatus(database, oldTask.ID, "In Progress", ""); err == nil {
		t.Fatal("historical task started under superseded approval")
	}
	newTasks, err := ListTasks(database, second.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(newTasks) != 1 || newTasks[0].Title != oldTask.Title {
		t.Fatalf("carried tasks = %#v", newTasks)
	}
}

func TestWaiverRequiresAttributionAndPreservesEvidenceHistory(t *testing.T) {
	database := openTestDatabase(t)
	item := readyItem(t, database, "Waiver")
	plan, err := CreatePlan(database, item.ID, "waiver")
	if err != nil {
		t.Fatal(err)
	}
	criterion, err := AddAcceptanceCriterionForPlan(database, plan.ID, "exception", "The exception is documented", "", true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AddTask(database, plan.ID, "Document exception", "", []string{"manual review"}); err != nil {
		t.Fatal(err)
	}
	if err := ApprovePlan(database, plan.ID); err != nil {
		t.Fatal(err)
	}
	if err := VerifyAcceptanceCriterion(database, criterion.ID, "Waived", "manual", "accepted"); err == nil {
		t.Fatal("waiver without attribution was accepted")
	}
	if err := VerifyAcceptanceCriterion(database, criterion.ID, "Waived", "manual", "accepted", "release-owner", "Legacy evidence is unavailable and risk was accepted"); err != nil {
		t.Fatal(err)
	}
	var evidenceCount int
	if err := database.QueryRow(`SELECT COUNT(*) FROM criterion_evidence WHERE acceptance_criterion_id=?`, criterion.ID).Scan(&evidenceCount); err != nil {
		t.Fatal(err)
	}
	if evidenceCount != 1 {
		t.Fatalf("evidence history count = %d", evidenceCount)
	}
}

func TestBlockedTaskCanResumeAndCompleteWithHistory(t *testing.T) {
	database := openTestDatabase(t)
	item := readyItem(t, database, "Resume")
	plan, err := CreatePlan(database, item.ID, "resume")
	if err != nil {
		t.Fatal(err)
	}
	criterion, err := AddAcceptanceCriterionForPlan(database, plan.ID, "complete", "The resumed task is verified", "", true)
	if err != nil {
		t.Fatal(err)
	}
	task, err := AddTask(database, plan.ID, "Resume work", "", []string{"go test ./..."})
	if err != nil {
		t.Fatal(err)
	}
	if err := ApprovePlan(database, plan.ID); err != nil {
		t.Fatal(err)
	}
	if err := SetTaskStatus(database, task.ID, "In Progress", ""); err != nil {
		t.Fatal(err)
	}
	if err := SetTaskStatus(database, task.ID, "Blocked", "Waiting for dependency"); err != nil {
		t.Fatal(err)
	}
	if err := ResumeTask(database, task.ID); err != nil {
		t.Fatal(err)
	}
	if err := SetTaskStatus(database, task.ID, "Done", "Dependency arrived and work passed"); err != nil {
		t.Fatal(err)
	}
	if err := VerifyAcceptanceCriterion(database, criterion.ID, "Passed", "go test", "verified"); err != nil {
		t.Fatal(err)
	}
	if err := SetItemStatus(database, item.ID, "In Progress"); err != nil {
		t.Fatal(err)
	}
	if err := SetItemStatus(database, item.ID, "Done"); err != nil {
		t.Fatal(err)
	}
	events, err := ListTaskEvents(database, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 4 || events[1].Status != "Blocked" || events[2].Outcome != "resumed" {
		t.Fatalf("task events = %#v", events)
	}
}

func TestLegacyMigrationLeavesMissingContractUnresolved(t *testing.T) {
	path := filepath.Join(t.TempDir(), DatabaseFileName)
	database, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	database.SetMaxOpenConns(1)
	defer database.Close()
	if _, err := database.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		t.Fatal(err)
	}
	names, err := migrations.Names()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names[:len(names)-2] {
		source, err := migrations.Files.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := database.Exec(string(source)); err != nil {
			t.Fatal(err)
		}
		version := filepath.Base(name[:len(name)-4])
		if _, err := database.Exec(`INSERT INTO schema_migrations(version,applied_at) VALUES(?,?)`, version, now()); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := database.Exec(`INSERT INTO categories(name,created_at) VALUES('Needed',?)`, now()); err != nil {
		t.Fatal(err)
	}
	result, err := database.Exec(`INSERT INTO roadmap_items(title,category_id,horizon,status,created_at,updated_at) VALUES('Legacy',1,'Now','In Progress',?,?)`, now(), now())
	if err != nil {
		t.Fatal(err)
	}
	itemID, _ := result.LastInsertId()
	if _, err := database.Exec(`INSERT INTO plan_revisions(roadmap_item_id,revision,content,status,created_at,approved_at) VALUES(?,1,'free form legacy plan','Approved',?,?)`, itemID, now(), now()); err != nil {
		t.Fatal(err)
	}
	if err := applyMigrations(database); err != nil {
		t.Fatal(err)
	}
	var status, title string
	if err := database.QueryRow(`SELECT c.status,c.title FROM acceptance_criteria c JOIN plan_revisions p ON p.id=c.plan_revision_id WHERE p.roadmap_item_id=?`, itemID).Scan(&status, &title); err != nil {
		t.Fatal(err)
	}
	if status != "Pending" || title == "" {
		t.Fatalf("legacy criterion = %q %q", status, title)
	}
}
