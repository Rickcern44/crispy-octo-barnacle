package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rickcern44/cassor/internal/config"
	"github.com/rickcern44/cassor/internal/project"
	"github.com/rickcern44/cassor/internal/store"
)

func TestBuildGeneratesDeterministicFeatureRoutes(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, project.StateDirectory)
	if err := os.Mkdir(state, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, sourceGuidesDirectory, "retired-guide"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"PRD.md", "CASSOR_PLAN_PACKET_SCHEMA.md", "CASSOR_RECOVERY.md", "CASSOR_CONTEXT_CONTRACT.md", "SDD_LITE_MIGRATION_POLICY.md", "GETTING_STARTED.md", "WORKFLOW_GUIDE.md", "RESUME_WORK.md", "ROADMAP_GUIDE.md"} {
		if err := os.WriteFile(filepath.Join(root, "docs", name), []byte("# "+name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "GETTING_STARTED.md"), []byte("[Product requirements](PRD.md)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := config.Write(config.Path(state), config.Config{Name: "Demo"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Initialize(filepath.Join(state, store.DatabaseFileName)); err != nil {
		t.Fatal(err)
	}
	database, err := store.Open(filepath.Join(state, store.DatabaseFileName))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	item, err := store.AddItem(database, "Document static routes", "Publish feature detail pages.", "Needed", "Now", "")
	if err != nil {
		t.Fatal(err)
	}
	featureType := "Capability"
	if _, err := store.UpdateItemDossier(database, item.ID, &featureType, nil); err != nil {
		t.Fatal(err)
	}
	related, err := store.AddItem(database, "Route search", "", "Needed", "Next", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.AddFeatureRelationship(database, related.ID, item.ID, "extends"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.AddFeatureChangeLink(database, related.ID, item.ID); err != nil {
		t.Fatal(err)
	}
	if err := store.TransitionItemStatus(database, item.ID, "Ready"); err != nil {
		t.Fatal(err)
	}
	plan, err := store.CreatePlan(database, item.ID, "static route implementation")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.AddTask(database, plan.ID, "Generate task history", "Expose this task in static data."); err != nil {
		t.Fatal(err)
	}
	first, err := Build(root, state)
	if err != nil {
		t.Fatal(err)
	}
	initial, err := os.ReadFile(first)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Build(root, state)
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := os.ReadFile(second)
	if err != nil {
		t.Fatal(err)
	}
	if string(initial) != string(repeated) {
		t.Fatal("Build() output changed without state changes")
	}
	if !strings.Contains(string(initial), "Document static routes") {
		t.Fatal("roadmap data does not include item title")
	}
	if !strings.Contains(string(initial), "Generate task history") {
		t.Fatal("roadmap data does not include feature task history")
	}
	if !strings.Contains(string(initial), "Capability") || !strings.Contains(string(initial), "extends") {
		t.Fatal("roadmap data does not include dossier fields and relationships")
	}
	if !strings.Contains(string(initial), "change_links") {
		t.Fatal("roadmap data does not include capability change links")
	}
	guide, err := os.ReadFile(filepath.Join(root, sourceGuidesDirectory, "product-requirements", "+page.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(guide), "PRD.md") {
		t.Fatal("generated guide route does not include source documentation")
	}
	contextGuide, err := os.ReadFile(filepath.Join(root, sourceGuidesDirectory, "context-contract", "+page.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contextGuide), "CASSOR_CONTEXT_CONTRACT.md") {
		t.Fatal("generated guide route does not include the context contract")
	}
	gettingStartedGuide, err := os.ReadFile(filepath.Join(root, sourceGuidesDirectory, "getting-started", "+page.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(gettingStartedGuide), "](PRD.md)") || !strings.Contains(string(gettingStartedGuide), "](/guides/product-requirements/)") {
		t.Fatal("generated guide route does not rewrite the PRD link")
	}
	if _, err := os.Stat(filepath.Join(root, sourceGuidesDirectory, "retired-guide")); !os.IsNotExist(err) {
		t.Fatal("generated guide route did not remove retired guide directory")
	}
}

func TestGenerateBuildsPortableRoadmapWithoutDocumentationWorkspace(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, project.StateDirectory)
	if err := os.Mkdir(state, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := config.Write(config.Path(state), config.Config{Name: "Greenfield Demo"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Initialize(filepath.Join(state, store.DatabaseFileName)); err != nil {
		t.Fatal(err)
	}
	database, err := store.Open(filepath.Join(state, store.DatabaseFileName))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	packet := store.PlanPacket{
		RoadmapItem:        store.PacketRoadmapItem{Title: "Portable roadmap", Description: "Build without Node.", Category: "Needed", Horizon: "Now"},
		Goal:               "Generate a portable site.",
		AcceptanceCriteria: []store.PacketCriterion{{ID: "portable", Title: "A bare repository builds a roadmap."}},
		Tasks:              []store.PacketTask{{Title: "Render fallback", Description: "Write standalone HTML.", Verification: []string{"go test ./internal/site"}}},
	}
	if _, err := store.RecordApprovedPlan(database, packet, "test approval"); err != nil {
		t.Fatal(err)
	}
	output, err := Generate(root, state)
	if err != nil {
		t.Fatal(err)
	}
	if output != filepath.Join(root, relativeOutput) {
		t.Fatalf("output = %q", output)
	}
	page, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"Greenfield Demo", "Portable roadmap", "Render fallback", "A bare repository builds a roadmap.", "Search roadmap"} {
		if !strings.Contains(string(page), expected) {
			t.Fatalf("portable page does not include %q", expected)
		}
	}
}
