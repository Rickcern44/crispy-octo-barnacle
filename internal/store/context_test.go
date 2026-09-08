package store

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestScopedContextContainsFreshSessionAssignment(t *testing.T) {
	database := openTestDatabase(t)
	recorded, err := RecordApprovedPlan(database, PlanPacket{
		RoadmapItem:        PacketRoadmapItem{Title: "Scoped item", Category: "Needed", Horizon: "Now"},
		Goal:               "Resume implementation",
		AcceptanceCriteria: []PacketCriterion{{ID: "behavior", Title: "Behavior is verified"}},
		Constraints:        []string{"Keep the API stable", "Use the existing SQLite store"},
		Tasks:              []PacketTask{{Title: "Implement behavior", Verification: []string{"go test ./internal/store"}}},
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
	if context.NextAction.ID != recorded.Tasks[0].ID || len(context.NextAction.Verification) != 1 {
		t.Fatalf("next action = %#v", context.NextAction)
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
