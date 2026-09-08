package store

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestPortableStateRoundTripIsDeterministic(t *testing.T) {
	source := openTestDatabase(t)
	recorded, err := RecordApprovedPlan(source, PlanPacket{
		RoadmapItem:        PacketRoadmapItem{Title: "Portable feature", Category: "Needed", Horizon: "Now"},
		Goal:               "Preserve state",
		AcceptanceCriteria: []PacketCriterion{{ID: "portable", Title: "The state round trips"}},
		Tasks:              []PacketTask{{Title: "Export state", Verification: []string{"go test ./..."}}},
	}, "approved")
	if err != nil {
		t.Fatal(err)
	}
	itemTwo, err := AddItem(source, "Related feature", "", "Recommended", "Next", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AddFeatureRelationship(source, itemTwo.ID, recorded.Plan.ItemID, "depends_on"); err != nil {
		t.Fatal(err)
	}
	capability, err := AddItem(source, "Authentication capability", "", "Needed", "Now", "")
	if err != nil {
		t.Fatal(err)
	}
	featureType := "Capability"
	if _, err := UpdateItemDossier(source, capability.ID, &featureType, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := AddFeatureChangeLink(source, itemTwo.ID, capability.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := AddDossierArtifact(source, capability.ID, "research", "discovery", "Capability finding", "source inspection", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := AddPhaseRecord(source, recorded.Plan.ItemID, "Verify", "round trip verified"); err != nil {
		t.Fatal(err)
	}
	if err := SetTaskStatus(source, recorded.Tasks[0].ID, "In Progress", ""); err != nil {
		t.Fatal(err)
	}
	if err := SetTaskStatus(source, recorded.Tasks[0].ID, "Blocked", "test backup"); err != nil {
		t.Fatal(err)
	}
	if err := VerifyAcceptanceCriterion(source, recorded.Criteria[0].ID, "Passed", "go test", "verified"); err != nil {
		t.Fatal(err)
	}
	if _, err := AddFeatureReport(source, FeatureReport{ItemID: recorded.Plan.ItemID, ExecutionMode: "single-agent", Roles: []string{"implementation"}, Verification: "passed"}); err != nil {
		t.Fatal(err)
	}
	export, err := ExportState(source)
	if err != nil {
		t.Fatal(err)
	}
	target := openTestDatabase(t)
	if err := ImportState(target, export); err != nil {
		t.Fatal(err)
	}
	roundTrip, err := ExportState(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(export, roundTrip) {
		t.Fatalf("export changed after round trip")
	}
}

func TestPortableImportRejectsInvalidAndConflictingStateWithoutMutation(t *testing.T) {
	source := openTestDatabase(t)
	recorded, err := RecordApprovedPlan(source, PlanPacket{
		RoadmapItem:        PacketRoadmapItem{Title: "Import feature", Category: "Needed", Horizon: "Now"},
		Goal:               "Test import failures",
		AcceptanceCriteria: []PacketCriterion{{ID: "valid", Title: "Input is valid"}},
		Tasks:              []PacketTask{{Title: "Import state", Verification: []string{"go test ./..."}}},
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	export, err := ExportState(source)
	if err != nil {
		t.Fatal(err)
	}
	target := openTestDatabase(t)
	if err := ImportState(target, []byte("not json")); err == nil {
		t.Fatal("invalid JSON was accepted")
	}
	var state PortableState
	if err := json.Unmarshal(export, &state); err != nil {
		t.Fatal(err)
	}
	state.Version = 99
	unknownVersion, _ := json.Marshal(state)
	if err := ImportState(target, unknownVersion); err == nil {
		t.Fatal("unknown format version was accepted")
	}
	if _, err := AddItem(target, "Existing", "", "Needed", "Now", ""); err != nil {
		t.Fatal(err)
	}
	if err := ImportState(target, export); err == nil {
		t.Fatal("non-empty destination was accepted")
	}
	var count int
	if err := target.QueryRow(`SELECT COUNT(*) FROM roadmap_items`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("destination changed after conflict: %d items", count)
	}
	_ = recorded
}

func TestPortableImportRollsBackConstraintFailure(t *testing.T) {
	source := openTestDatabase(t)
	if _, err := RecordApprovedPlan(source, PlanPacket{
		RoadmapItem:        PacketRoadmapItem{Title: "Rollback feature", Category: "Needed", Horizon: "Now"},
		Goal:               "Test rollback",
		AcceptanceCriteria: []PacketCriterion{{ID: "rollback", Title: "Import rolls back"}},
		Tasks:              []PacketTask{{Title: "reject", Verification: []string{"go test ./..."}}},
	}, ""); err != nil {
		t.Fatal(err)
	}
	export, err := ExportState(source)
	if err != nil {
		t.Fatal(err)
	}
	target := openTestDatabase(t)
	if _, err := target.Exec(`CREATE TRIGGER reject_import_task BEFORE INSERT ON tasks WHEN NEW.title='reject' BEGIN SELECT RAISE(ABORT, 'test import failure'); END`); err != nil {
		t.Fatal(err)
	}
	if err := ImportState(target, export); err == nil {
		t.Fatal("trigger failure was accepted")
	}
	for _, table := range []string{"roadmap_items", "plan_revisions", "acceptance_criteria", "tasks"} {
		var count int
		if err := target.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("%s count = %d after import rollback", table, count)
		}
	}
}
