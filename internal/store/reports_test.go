package store

import "testing"

func TestFeatureReportContextMetricsPreserveObservedZeroAndUnavailable(t *testing.T) {
	database := openTestDatabase(t)
	item, err := AddItem(database, "Report feature", "", "Needed", "Now", "")
	if err != nil {
		t.Fatal(err)
	}
	key := "baseline"
	zero := int64(0)
	saved, err := AddFeatureReport(database, FeatureReport{ItemID: item.ID, ExecutionMode: "single-agent", ComparisonKey: &key, ContextPackets: &zero, ContextBytes: &zero, Handoffs: &zero})
	if err != nil {
		t.Fatal(err)
	}
	if saved.CreatedAt == "" || saved.ID == 0 {
		t.Fatalf("saved report missing identity: %+v", saved)
	}
	values, err := ListFeatureReports(database)
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 1 {
		t.Fatalf("reports = %d, want 1", len(values))
	}
	got := values[0]
	if got.ComparisonKey == nil || *got.ComparisonKey != key {
		t.Fatalf("comparison key = %v, want %q", got.ComparisonKey, key)
	}
	for name, value := range map[string]*int64{"context_packets": got.ContextPackets, "context_bytes": got.ContextBytes, "handoffs": got.Handoffs} {
		if value == nil || *value != 0 {
			t.Fatalf("%s = %v, want pointer to zero", name, value)
		}
	}

	if _, err := AddFeatureReport(database, FeatureReport{ItemID: item.ID, ExecutionMode: "multi-agent"}); err != nil {
		t.Fatal(err)
	}
	values, err = ListFeatureReports(database)
	if err != nil {
		t.Fatal(err)
	}
	if values[1].ComparisonKey != nil {
		t.Fatalf("comparison_key = %v, want NULL", values[1].ComparisonKey)
	}
	for name, value := range map[string]*int64{"context_packets": values[1].ContextPackets, "context_bytes": values[1].ContextBytes, "handoffs": values[1].Handoffs} {
		if value != nil {
			t.Fatalf("%s = %v, want NULL", name, value)
		}
	}
}
