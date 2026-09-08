package store

import "testing"

func TestDossierChangeLinksArtifactsAndAcceptedState(t *testing.T) {
	database := openTestDatabase(t)
	capability, err := AddItem(database, "Authentication", "", "Needed", "Now", "")
	if err != nil {
		t.Fatal(err)
	}
	featureType := "Capability"
	if _, err := UpdateItemDossier(database, capability.ID, &featureType, nil); err != nil {
		t.Fatal(err)
	}
	change, err := AddItem(database, "Passkeys", "", "Needed", "Next", "")
	if err != nil {
		t.Fatal(err)
	}
	link, err := AddFeatureChangeLink(database, change.ID, capability.ID)
	if err != nil {
		t.Fatal(err)
	}
	if link.ChangeItemID != change.ID || link.CapabilityItemID != capability.ID {
		t.Fatalf("link = %#v", link)
	}
	if err := TransitionItemStatus(database, change.ID, "Ready"); err != nil {
		t.Fatal(err)
	}
	if err := TransitionItemStatus(database, change.ID, "In Progress"); err != nil {
		t.Fatal(err)
	}
	if err := TransitionItemStatus(database, change.ID, "Done"); err != nil {
		t.Fatal(err)
	}
	if err := AcceptCapabilityState(database, capability.ID, change.ID, "Password login and session renewal are supported.", "orchestrator"); err != nil {
		t.Fatal(err)
	}
	updated, err := GetItem(database, capability.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.CurrentState == "" {
		t.Fatal("accepted capability state was not projected")
	}
	history, err := ListCapabilityStateHistoryForItem(database, capability.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].SourceChangeID == nil || *history[0].SourceChangeID != change.ID {
		t.Fatalf("state history = %#v", history)
	}
	secondChange, err := AddItem(database, "Session controls", "", "Needed", "Next", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AddFeatureChangeLink(database, secondChange.ID, capability.ID); err != nil {
		t.Fatal(err)
	}
	for _, status := range []string{"Ready", "In Progress", "Done"} {
		if err := TransitionItemStatus(database, secondChange.ID, status); err != nil {
			t.Fatal(err)
		}
	}
	if err := AcceptCapabilityState(database, capability.ID, secondChange.ID, "Session controls and renewal limits are supported.", "orchestrator"); err != nil {
		t.Fatal(err)
	}
	updated, err = GetItem(database, capability.ID)
	if err != nil {
		t.Fatal(err)
	}
	history, err = ListCapabilityStateHistoryForItem(database, capability.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.CurrentState != "Session controls and renewal limits are supported." || len(history) != 2 || history[0].State == history[1].State {
		t.Fatalf("successive capability state = item %#v history %#v", updated, history)
	}
	draft, err := AddDossierArtifact(database, capability.ID, "research", "discovery", "Session renewal behavior was confirmed.", "repository inspection", nil)
	if err != nil {
		t.Fatal(err)
	}
	if draft.Status != "Draft" {
		t.Fatalf("draft status = %q", draft.Status)
	}
	beforeDraft, err := GetItem(database, capability.ID)
	if err != nil {
		t.Fatal(err)
	}
	if beforeDraft.CurrentState != "Session controls and renewal limits are supported." {
		t.Fatal("draft artifact changed canonical capability state")
	}
	if err := AcceptDossierArtifact(database, draft.ID, "orchestrator"); err != nil {
		t.Fatal(err)
	}
	artifacts, err := ListDossierArtifactsForItem(database, capability.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(artifacts) != 1 || artifacts[0].Status != "Accepted" {
		t.Fatalf("artifacts = %#v", artifacts)
	}
	if _, err := UpdateItemDossier(database, capability.ID, nil, stringPointer("unauthorized state")); err == nil {
		t.Fatal("direct current-state update bypassed acceptance")
	}
}

func TestDossierRejectsUnverifiedStateAndEvidenceFreeAcceptance(t *testing.T) {
	database := openTestDatabase(t)
	capability, err := AddItem(database, "Authentication", "", "Needed", "Now", "")
	if err != nil {
		t.Fatal(err)
	}
	featureType := "Capability"
	if _, err := UpdateItemDossier(database, capability.ID, &featureType, nil); err != nil {
		t.Fatal(err)
	}
	change, err := AddItem(database, "Passkeys", "", "Needed", "Next", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AddFeatureChangeLink(database, change.ID, capability.ID); err != nil {
		t.Fatal(err)
	}
	if err := AcceptCapabilityState(database, capability.ID, change.ID, "not verified", "orchestrator"); err == nil {
		t.Fatal("accepted state from incomplete change")
	}
	draft, err := AddDossierArtifact(database, capability.ID, "research", "discovery", "Unverified finding", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := AcceptDossierArtifact(database, draft.ID, "orchestrator"); err == nil {
		t.Fatal("accepted an artifact without evidence")
	}
}
