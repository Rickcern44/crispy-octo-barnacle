package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rickcern44/cassor/internal/store"
)

func TestPlanTemplateContainsItemIDAndValidates(t *testing.T) {
	_, database := workflowTestProject(t)
	defer database.Close()
	item, err := store.AddItem(database, "Template item", "Description", "Needed", "Now", "Rationale")
	if err != nil {
		t.Fatal(err)
	}

	command := NewRootCommand()
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetArgs([]string{"plan", "template", "--item", "1"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"id":1`) {
		t.Fatalf("template = %q", output.String())
	}
	path := filepath.Join(t.TempDir(), "template.json")
	if err := os.WriteFile(path, output.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}

	validate := NewRootCommand()
	var validation bytes.Buffer
	validate.SetOut(&validation)
	validate.SetArgs([]string{"plan", "validate", "--file", path})
	if err := validate.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(validation.String()); got != "Plan packet valid for item 1" {
		t.Fatalf("validation output = %q", got)
	}
	if item.ID != 1 {
		t.Fatalf("item = %#v", item)
	}
}

func TestPlanRecordRejectsOpenQuestionsByDefaultWithoutWriting(t *testing.T) {
	_, database := workflowTestProject(t)
	database.Close()
	packet := store.PlanPacket{RoadmapItem: store.PacketRoadmapItem{Title: "Questions", Category: "Needed", Horizon: "Now"}, Goal: "Record safely", OpenQuestions: []string{"Which runtime?"}, AcceptanceCriteria: []store.PacketCriterion{{Title: "Recorded"}}, Tasks: []store.PacketTask{{Title: "Record", Verification: []string{"go test ./..."}}}}
	contents, err := json.Marshal(packet)
	if err != nil {
		t.Fatal(err)
	}
	path := writePlanPacket(t, string(contents))
	command := NewRootCommand()
	command.SetArgs([]string{"plan", "record", "--file", path, "--approve", "--approved-by-user"})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "unresolved open questions") {
		t.Fatalf("record error = %v", err)
	}
	check, err := store.Open(filepath.Join(".cassor", store.DatabaseFileName))
	if err != nil {
		t.Fatal(err)
	}
	defer check.Close()
	var count int
	if err := check.QueryRow(`SELECT COUNT(*) FROM plan_revisions`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("plan count = %d, want 0", count)
	}
}

func TestPlanRecordOpenQuestionOverrideWarnsAndPreservesContent(t *testing.T) {
	_, database := workflowTestProject(t)
	database.Close()
	packet := store.PlanPacket{RoadmapItem: store.PacketRoadmapItem{Title: "Questions", Category: "Needed", Horizon: "Now"}, Goal: "Record explicitly", OpenQuestions: []string{"Which runtime?"}, AcceptanceCriteria: []store.PacketCriterion{{Title: "Recorded"}}, Tasks: []store.PacketTask{{Title: "Record", Verification: []string{"go test ./..."}}}}
	contents, err := json.Marshal(packet)
	if err != nil {
		t.Fatal(err)
	}
	path := writePlanPacket(t, string(contents))
	command := NewRootCommand()
	var stderr bytes.Buffer
	command.SetErr(&stderr)
	command.SetArgs([]string{"plan", "record", "--file", path, "--approve", "--approved-by-user", "--allow-open-questions"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr.String(), "WARNING") {
		t.Fatalf("warning = %q", stderr.String())
	}
	check, err := store.Open(filepath.Join(".cassor", store.DatabaseFileName))
	if err != nil {
		t.Fatal(err)
	}
	defer check.Close()
	var raw string
	if err := check.QueryRow(`SELECT content FROM plan_revisions`).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var recorded store.PlanPacket
	if err := json.Unmarshal([]byte(raw), &recorded); err != nil {
		t.Fatal(err)
	}
	if len(recorded.OpenQuestions) != 1 || recorded.OpenQuestions[0] != "Which runtime?" {
		t.Fatalf("recorded questions = %#v", recorded.OpenQuestions)
	}
}

func TestPlanValidateRejectsMalformedInputWithoutDatabaseChanges(t *testing.T) {
	_, database := workflowTestProject(t)
	defer database.Close()
	if _, err := store.AddItem(database, "Existing item", "", "Needed", "Now", ""); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "malformed.json")
	if err := os.WriteFile(path, []byte(`{"goal":"missing required arrays"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var beforeItems, beforePlans int
	if err := database.QueryRow(`SELECT COUNT(*) FROM roadmap_items`).Scan(&beforeItems); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow(`SELECT COUNT(*) FROM plan_revisions`).Scan(&beforePlans); err != nil {
		t.Fatal(err)
	}

	command := NewRootCommand()
	command.SetArgs([]string{"plan", "validate", "--file", path})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "validate plan packet") {
		t.Fatalf("malformed validation error = %v", err)
	}
	var afterItems, afterPlans int
	if err := database.QueryRow(`SELECT COUNT(*) FROM roadmap_items`).Scan(&afterItems); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow(`SELECT COUNT(*) FROM plan_revisions`).Scan(&afterPlans); err != nil {
		t.Fatal(err)
	}
	if beforeItems != afterItems || beforePlans != afterPlans {
		t.Fatalf("database counts changed: before items/plans=%d/%d after=%d/%d", beforeItems, beforePlans, afterItems, afterPlans)
	}
}
