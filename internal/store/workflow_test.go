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
	if _, err := AddAcceptanceCriterionForPlan(database, plan.ID, "contract", "The work is complete", "", true); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`UPDATE tasks SET verification='["go test ./..."]' WHERE id=?`, task.ID); err != nil {
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

func TestItemStatusTransitionsAreGuarded(t *testing.T) {
	path := filepath.Join(t.TempDir(), DatabaseFileName)
	if err := Initialize(path); err != nil {
		t.Fatal(err)
	}
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	item, err := AddItem(database, "Ship lifecycle", "", "Needed", "Now", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := TransitionItemStatus(database, item.ID, "Done"); err == nil {
		t.Fatal("TransitionItemStatus() completed a proposed item")
	}
	for _, status := range []string{"Ready", "In Progress", "Done"} {
		if err := TransitionItemStatus(database, item.ID, status); err != nil {
			t.Fatalf("TransitionItemStatus(%q): %v", status, err)
		}
	}
	if err := TransitionItemStatus(database, item.ID, "In Progress"); err == nil {
		t.Fatal("TransitionItemStatus() restarted a completed item")
	}
}

func TestCompletionCandidatesRequireEveryApprovedTaskAndAnOpenItem(t *testing.T) {
	path := filepath.Join(t.TempDir(), DatabaseFileName)
	if err := Initialize(path); err != nil {
		t.Fatal(err)
	}
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	completed, err := AddItem(database, "Completed plan", "", "Needed", "Now", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := SetItemStatus(database, completed.ID, "Ready"); err != nil {
		t.Fatal(err)
	}
	completedPlan, err := CreatePlan(database, completed.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	completedTask, err := AddTask(database, completedPlan.ID, "Done task", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AddAcceptanceCriterionForPlan(database, completedPlan.ID, "done", "The task is complete", "", true); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`UPDATE tasks SET verification='["go test ./..."]' WHERE id=?`, completedTask.ID); err != nil {
		t.Fatal(err)
	}
	if err := ApprovePlan(database, completedPlan.ID); err != nil {
		t.Fatal(err)
	}
	if err := SetTaskStatus(database, completedTask.ID, "In Progress", ""); err != nil {
		t.Fatal(err)
	}
	if err := SetTaskStatus(database, completedTask.ID, "Done", "done"); err != nil {
		t.Fatal(err)
	}
	criteria, err := ListAcceptanceCriteria(database)
	if err != nil || len(criteria) != 1 {
		t.Fatalf("criteria = %#v, err = %v", criteria, err)
	}
	if err := VerifyAcceptanceCriterion(database, criteria[0].ID, "Passed", "go test", "verified"); err != nil {
		t.Fatal(err)
	}

	incomplete, err := AddItem(database, "Incomplete plan", "", "Needed", "Now", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := SetItemStatus(database, incomplete.ID, "Ready"); err != nil {
		t.Fatal(err)
	}
	incompletePlan, err := CreatePlan(database, incomplete.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AddTask(database, incompletePlan.ID, "Open task", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := AddAcceptanceCriterionForPlan(database, incompletePlan.ID, "done", "The task is complete", "", true); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`UPDATE tasks SET verification='["go test ./..."]' WHERE plan_revision_id=?`, incompletePlan.ID); err != nil {
		t.Fatal(err)
	}
	if err := ApprovePlan(database, incompletePlan.ID); err != nil {
		t.Fatal(err)
	}

	candidates, err := CompletionCandidates(database)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 1 || candidates[0].ItemID != completed.ID || candidates[0].TaskCount != 1 {
		t.Fatalf("candidates = %#v", candidates)
	}
	if err := SetItemStatus(database, completed.ID, "In Progress"); err != nil {
		t.Fatal(err)
	}
	if err := SetItemStatus(database, completed.ID, "Done"); err != nil {
		t.Fatal(err)
	}
	candidates, err = CompletionCandidates(database)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 0 {
		t.Fatalf("completed item remained a candidate: %#v", candidates)
	}
}

func TestFeatureDossierFieldsAndRelationshipsAreValidated(t *testing.T) {
	path := filepath.Join(t.TempDir(), DatabaseFileName)
	if err := Initialize(path); err != nil {
		t.Fatal(err)
	}
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	capability, err := AddItem(database, "Authentication", "", "Needed", "Now", "")
	if err != nil {
		t.Fatal(err)
	}
	change, err := AddItem(database, "Passkeys", "", "Needed", "Next", "")
	if err != nil {
		t.Fatal(err)
	}
	if capability.FeatureType != "Change" || capability.CurrentState != "" {
		t.Fatalf("new item dossier defaults = %#v", capability)
	}
	featureType, currentState := "Capability", "Password login and session renewal are supported."
	updated, err := UpdateItemDossier(database, capability.ID, &featureType, &currentState)
	if err != nil {
		t.Fatal(err)
	}
	if updated.FeatureType != featureType || updated.CurrentState != currentState {
		t.Fatalf("dossier update = %#v", updated)
	}
	invalidType := "Unknown"
	if _, err := UpdateItemDossier(database, capability.ID, &invalidType, nil); err == nil {
		t.Fatal("UpdateItemDossier() accepted an invalid feature type")
	}
	relationship, err := AddFeatureRelationship(database, change.ID, capability.ID, "extends")
	if err != nil {
		t.Fatal(err)
	}
	if relationship.SourceItemTitle != "Passkeys" || relationship.TargetItemTitle != "Authentication" {
		t.Fatalf("relationship = %#v", relationship)
	}
	if _, err := AddFeatureRelationship(database, change.ID, capability.ID, "extends"); err == nil {
		t.Fatal("AddFeatureRelationship() accepted a duplicate")
	}
	if _, err := AddFeatureRelationship(database, change.ID, change.ID, "extends"); err == nil {
		t.Fatal("AddFeatureRelationship() accepted a self-link")
	}
	if _, err := AddFeatureRelationship(database, change.ID, capability.ID, "related_to"); err == nil {
		t.Fatal("AddFeatureRelationship() accepted an invalid relationship type")
	}
	relationships, err := ListFeatureRelationships(database, capability.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(relationships) != 1 || relationships[0].ID != relationship.ID {
		t.Fatalf("relationships = %#v", relationships)
	}
}
