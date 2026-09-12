package store

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestScopedContextContainsFreshSessionAssignment(t *testing.T) {
	database := openTestDatabase(t)
	recorded, err := RecordApprovedPlan(database, PlanPacket{
		RoadmapItem:        PacketRoadmapItem{Title: "Scoped item", Category: "Needed", Horizon: "Now"},
		Goal:               "Resume implementation",
		Scope:              PacketScope{Included: []string{"Context enrichment"}, Excluded: []string{"Unrelated redesign"}},
		Decisions:          []string{"Keep the API stable"},
		AcceptanceCriteria: []PacketCriterion{{ID: "behavior", Title: "Behavior is verified"}},
		Constraints:        []string{"Keep the API stable", "Use the existing SQLite store"},
		Tasks:              []PacketTask{{Title: "Implement behavior", Description: "Populate the active task context", Verification: []string{"go test ./internal/store"}}, {Title: "Unrelated follow-up", Description: "Do not include this description", Verification: []string{"go test ./internal/store"}}},
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	context, err := BuildScopedContext(database, recorded.Plan.ItemID, "implementation", 12000)
	if err != nil {
		t.Fatal(err)
	}
	if context.Assignment.PlanID != recorded.Plan.ID || context.Assignment.Goal != "Resume implementation" {
		t.Fatalf("assignment = %#v", context.Assignment)
	}
	if len(context.Constraints) != 2 || len(context.Criteria) != 1 || context.Criteria[0].Status != "Pending" {
		t.Fatalf("scoped requirements = %#v / %#v", context.Constraints, context.Criteria)
	}
	if len(context.Scope.Included) != 1 || context.Scope.Included[0] != "Context enrichment" || len(context.Scope.Excluded) != 1 || context.Scope.Excluded[0] != "Unrelated redesign" || len(context.Decisions) != 1 || context.Decisions[0] != "Keep the API stable" {
		t.Fatalf("approved plan context = scope %#v, decisions %#v", context.Scope, context.Decisions)
	}
	if context.NextAction.ID != recorded.Tasks[0].ID || context.NextAction.Description != "Populate the active task context" || len(context.NextAction.Verification) != 1 {
		t.Fatalf("next action = %#v", context.NextAction)
	}
	encodedContext, err := json.Marshal(context)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encodedContext), "Do not include this description") {
		t.Fatal("unrelated task description leaked into scoped context")
	}
	encoded, err := json.Marshal(context)
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) > 12000 || context.Bytes != len(encoded) {
		t.Fatalf("encoded bytes = %d, recorded = %d", len(encoded), context.Bytes)
	}
}

func TestScopedContextIgnoresUnrelatedBacklog(t *testing.T) {
	database := openTestDatabase(t)
	recorded, err := RecordApprovedPlan(database, PlanPacket{
		RoadmapItem:        PacketRoadmapItem{Title: "Stable item", Category: "Needed", Horizon: "Now"},
		Goal:               "Keep output stable",
		AcceptanceCriteria: []PacketCriterion{{ID: "stable", Title: "Output is stable"}},
		Tasks:              []PacketTask{{Title: "Measure output", Verification: []string{"go test ./..."}}},
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	before, err := BuildScopedContext(database, recorded.Plan.ItemID, "implementation", 12000)
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 100; index++ {
		if _, err := AddItem(database, strings.Repeat("unrelated ", 20), "large backlog", "Needed", "Later", ""); err != nil {
			t.Fatal(err)
		}
	}
	after, err := BuildScopedContext(database, recorded.Plan.ItemID, "implementation", 12000)
	if err != nil {
		t.Fatal(err)
	}
	if before.Bytes != after.Bytes || before.Assignment.PlanID != after.Assignment.PlanID || len(after.Omitted) != len(before.Omitted) {
		t.Fatalf("scoped output changed with unrelated backlog: before=%d after=%d", before.Bytes, after.Bytes)
	}
}

func TestScopedContextSignalsTruncationAndStableReferences(t *testing.T) {
	database := openTestDatabase(t)
	criteria := make([]PacketCriterion, 0, 8)
	tasks := make([]PacketTask, 0, 8)
	for index := 0; index < 8; index++ {
		criteria = append(criteria, PacketCriterion{ID: "criterion-" + string(rune('a'+index)), Title: strings.Repeat("criterion detail ", 10)})
		tasks = append(tasks, PacketTask{Title: strings.Repeat("task detail ", 10), Verification: []string{"go test ./..."}})
	}
	recorded, err := RecordApprovedPlan(database, PlanPacket{RoadmapItem: PacketRoadmapItem{Title: "Large item", Category: "Needed", Horizon: "Now"}, Goal: strings.Repeat("goal ", 500), AcceptanceCriteria: criteria, Tasks: tasks}, "")
	if err != nil {
		t.Fatal(err)
	}
	context, err := BuildScopedContext(database, recorded.Plan.ItemID, "implementation", 6000)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(context)
	if err != nil {
		t.Fatal(err)
	}
	if !context.Truncated || len(context.Omitted) == 0 || len(encoded) > 6000 {
		t.Fatalf("truncation = %t, omitted = %d, bytes = %d", context.Truncated, len(context.Omitted), len(encoded))
	}
	for _, reference := range context.Omitted {
		if reference.ID == 0 || reference.Command == "" {
			t.Fatalf("unstable omitted reference = %#v", reference)
		}
	}
}

func TestScopedContextBoundsStructuredDetailsWithStableReferences(t *testing.T) {
	database := openTestDatabase(t)
	recorded, err := RecordApprovedPlan(database, PlanPacket{
		RoadmapItem:        PacketRoadmapItem{Title: "Structured context", Category: "Needed", Horizon: "Now"},
		Goal:               strings.Repeat("goal ", 300),
		Scope:              PacketScope{Included: []string{strings.Repeat("included ", 120)}, Excluded: []string{strings.Repeat("excluded ", 120)}},
		Decisions:          []string{strings.Repeat("decision ", 120)},
		AcceptanceCriteria: []PacketCriterion{{ID: "verify", Title: "Verification"}},
		Tasks:              []PacketTask{{Title: "Resume work", Description: strings.Repeat("description ", 200), Verification: []string{"go test ./internal/store/..."}}},
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	first, err := BuildScopedContext(database, recorded.Plan.ItemID, "implementation", 2200)
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildScopedContext(database, recorded.Plan.ItemID, "implementation", 2200)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Truncated || len(encoded) > 2200 || first.Bytes != len(encoded) {
		t.Fatalf("bounded context = truncated %t, bytes %d/%d", first.Truncated, len(encoded), first.Bytes)
	}
	if first.Bytes != second.Bytes || len(first.Omitted) != len(second.Omitted) {
		t.Fatalf("omissions are not stable: %#v / %#v", first.Omitted, second.Omitted)
	}
	var foundPlan, foundTask bool
	for _, reference := range first.Omitted {
		switch reference.Kind {
		case "decision-detail", "scope-detail":
			if reference.Command != "cassor plan show "+fmt.Sprint(recorded.Plan.ID) {
				t.Fatalf("structured omission reference = %#v", reference)
			}
			foundPlan = true
		case "next-action-description":
			if reference.Command != "cassor task show "+fmt.Sprint(recorded.Tasks[0].ID) {
				t.Fatalf("description omission reference = %#v", reference)
			}
			foundTask = true
		}
	}
	if !foundPlan || !foundTask || first.NextAction.Description != "" || first.NextAction.ID != recorded.Tasks[0].ID {
		t.Fatalf("structured omissions = %#v; next action = %#v", first.Omitted, first.NextAction)
	}
}

func TestDefaultContextIsNavigationBoundedAndRolesAreExplicit(t *testing.T) {
	database := openTestDatabase(t)
	for index := 0; index < 25; index++ {
		if _, err := AddItem(database, "Planned item "+string(rune('a'+index)), strings.Repeat("detail ", 100), "Needed", "Now", ""); err != nil {
			t.Fatal(err)
		}
	}
	context, err := CompactContext(database)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(context)
	if err != nil {
		t.Fatal(err)
	}
	if context.Counts.ProposedItems != 25 || len(context.ProposedItems) != 10 || !context.Truncated || strings.Contains(string(encoded), "detail detail") {
		t.Fatalf("default context = %#v", context)
	}
	if _, err := BuildScopedContext(database, 1, "unknown", 12000); err == nil {
		t.Fatal("unknown context role was accepted")
	}
}

func TestDefaultContextProjectsCapabilitiesChangesAndGaps(t *testing.T) {
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
	if err := TransitionItemStatus(database, change.ID, "Ready"); err != nil {
		t.Fatal(err)
	}
	gap, err := AddItem(database, "Legacy session gap", "", "Technical debt", "Later", "")
	if err != nil {
		t.Fatal(err)
	}
	gapType := "Gap"
	if _, err := UpdateItemDossier(database, gap.ID, &gapType, nil); err != nil {
		t.Fatal(err)
	}
	context, err := CompactContext(database)
	if err != nil {
		t.Fatal(err)
	}
	if context.Counts.Capabilities != 1 || context.Counts.PlannedWork != 1 || context.Counts.KnownGaps != 1 {
		t.Fatalf("dossier projections = %#v", context.Counts)
	}
}

func TestContextProjectsEpicParentAndNavigation(t *testing.T) {
	database := openTestDatabase(t)
	epic, err := AddEpic(database, "Onboarding", "Account setup work")
	if err != nil {
		t.Fatal(err)
	}
	recorded, err := RecordApprovedPlan(database, PlanPacket{
		RoadmapItem:        PacketRoadmapItem{Title: "Guided setup", Category: "Needed", Horizon: "Now", EpicID: &epic.ID},
		Goal:               "Guide setup",
		AcceptanceCriteria: []PacketCriterion{{Title: "Setup is guided"}},
		Tasks:              []PacketTask{{Title: "Implement setup", Verification: []string{"go test ./internal/store"}}},
	}, "approved")
	if err != nil {
		t.Fatal(err)
	}
	compact, err := CompactContext(database)
	if err != nil {
		t.Fatal(err)
	}
	if compact.Counts.Epics != 1 || len(compact.Epics) != 1 || compact.Epics[0].Command != "cassor epic show 1" {
		t.Fatalf("Epic navigation = %#v", compact)
	}
	scoped, err := BuildScopedContext(database, recorded.Plan.ItemID, "implementation", 12000)
	if err != nil {
		t.Fatal(err)
	}
	if scoped.Epic == nil || scoped.Epic.ID != epic.ID || scoped.Epic.Title != epic.Title {
		t.Fatalf("Epic parent = %#v", scoped.Epic)
	}
}
