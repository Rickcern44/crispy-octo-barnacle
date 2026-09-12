package store

import (
	"encoding/json"
	"testing"
)

func TestEpicParentLifecycle(t *testing.T) {
	database := openTestDatabase(t)
	epic, err := AddEpic(database, "Payments", "Payment outcomes")
	if err != nil {
		t.Fatal(err)
	}
	item := readyItem(t, database, "Card payments")
	assigned, err := SetItemEpic(database, item.ID, &epic.ID)
	if err != nil {
		t.Fatal(err)
	}
	if assigned.EpicID == nil || *assigned.EpicID != epic.ID {
		t.Fatalf("epic parent = %#v", assigned.EpicID)
	}
	if err := DeleteEpic(database, epic.ID); err == nil {
		t.Fatal("DeleteEpic accepted an Epic with children")
	}
	if _, err := SetItemEpic(database, item.ID, nil); err != nil {
		t.Fatal(err)
	}
	if err := DeleteEpic(database, epic.ID); err != nil {
		t.Fatal(err)
	}
}

func TestEpicPortableRoundTripAndPacketCompatibility(t *testing.T) {
	source := openTestDatabase(t)
	epic, err := AddEpic(source, "Search", "Search improvements")
	if err != nil {
		t.Fatal(err)
	}
	recorded, err := RecordApprovedPlan(source, PlanPacket{
		RoadmapItem: PacketRoadmapItem{Title: "Indexed search", Category: "Needed", Horizon: "Now", EpicID: &epic.ID},
		Goal:        "Improve search", AcceptanceCriteria: []PacketCriterion{{Title: "Search works"}}, Tasks: []PacketTask{{Title: "Implement", Verification: []string{"go test ./..."}}},
	}, "approved")
	if err != nil {
		t.Fatal(err)
	}
	if recorded.Plan.ItemID == 0 {
		t.Fatal("missing recorded item")
	}
	export, err := ExportState(source)
	if err != nil {
		t.Fatal(err)
	}
	target := openTestDatabase(t)
	if err := ImportState(target, export); err != nil {
		t.Fatal(err)
	}
	item, err := GetItem(target, recorded.Plan.ItemID)
	if err != nil || item.EpicID == nil || *item.EpicID != epic.ID {
		t.Fatalf("imported parent = %#v, error = %v", item.EpicID, err)
	}
	var state PortableState
	if err := json.Unmarshal(export, &state); err != nil {
		t.Fatal(err)
	}
	state.Epics = nil
	for index := range state.Items {
		state.Items[index].EpicID = nil
	}
	legacy, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	legacyTarget := openTestDatabase(t)
	if err := ImportState(legacyTarget, legacy); err != nil {
		t.Fatal(err)
	}
}
