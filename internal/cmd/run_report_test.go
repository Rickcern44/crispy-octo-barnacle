package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rickcern44/cassor/internal/store"
)

func TestRunReportMarksUnavailableTokenMetrics(t *testing.T) {
	command := newRunReportCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"--mode", "parallel", "--roles", "discovery,verification", "--elapsed", "42s", "--tool-calls", "5", "--verification", "passed"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Tokens: input=unavailable") {
		t.Fatalf("report = %q", output.String())
	}
}

func TestRunReportRendersObservedComparisonMetrics(t *testing.T) {
	command := newRunReportCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"--mode", "parallel", "--elapsed", "1s", "--comparison-key", "baseline", "--context-packets", "0", "--context-bytes", "12", "--handoffs", "2"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	for _, want := range []string{"Comparison key: baseline", "Context packets: 0", "Context bytes: 12", "Handoffs: 2"} {
		if !strings.Contains(text, want) {
			t.Fatalf("report = %q, missing %q", text, want)
		}
	}
}

func TestRunReportRecordAndCompareFiltersByKey(t *testing.T) {
	_, database := workflowTestProject(t)
	defer database.Close()
	item, err := store.AddItem(database, "Comparable feature", "", "Needed", "Now", "")
	if err != nil {
		t.Fatal(err)
	}

	record := NewRootCommand()
	record.SetOut(new(bytes.Buffer))
	record.SetArgs([]string{"run-report", "record", "--item", "1", "--mode", "single-agent", "--roles", "implementation", "--elapsed", "2s", "--comparison-key", "baseline", "--context-packets", "0", "--handoffs", "0"})
	if err := record.Execute(); err != nil {
		t.Fatal(err)
	}
	otherKey := "other"
	zero := int64(1)
	if _, err := store.AddFeatureReport(database, store.FeatureReport{ItemID: item.ID, ExecutionMode: "parallel", ComparisonKey: &otherKey, ContextPackets: &zero}); err != nil {
		t.Fatal(err)
	}

	compare := NewRootCommand()
	output := new(bytes.Buffer)
	compare.SetOut(output)
	compare.SetArgs([]string{"run-report", "compare", "--comparison-key", "baseline"})
	if err := compare.Execute(); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	if !strings.Contains(text, "Comparable feature") || !strings.Contains(text, "single-agent") || !strings.Contains(text, "0") {
		t.Fatalf("comparison = %q", text)
	}
	if strings.Contains(text, "parallel") {
		t.Fatalf("comparison included different key: %q", text)
	}
}

func TestRunReportJSONIncludesObservedTokens(t *testing.T) {
	command := newRunReportCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"--mode", "single-agent", "--elapsed", "1m", "--input-tokens", "120", "--total-tokens", "200", "--json"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "\"input_tokens\": 120") || !strings.Contains(output.String(), "\"total_tokens\": 200") {
		t.Fatalf("json report = %q", output.String())
	}
}
