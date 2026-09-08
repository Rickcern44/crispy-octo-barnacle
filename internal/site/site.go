// Package site renders Cassor's static developer documentation source.
package site

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rickcern44/cassor/internal/config"
	"github.com/rickcern44/cassor/internal/store"
)

const (
	sourceDocsDirectory   = "docs-site/src/lib/generated"
	sourceGuidesDirectory = "docs-site/src/routes/guides"
	relativeOutput        = "docs/roadmap/index.html"
)

type roadmapData struct {
	GeneratedAt      string                        `json:"generated_at"`
	ProjectName      string                        `json:"project_name"`
	Items            []store.Item                  `json:"items"`
	Plans            []planView                    `json:"plans"`
	Tasks            []store.Task                  `json:"tasks"`
	Reports          []store.FeatureReport         `json:"reports"`
	Phases           []store.PhaseRecord           `json:"phases"`
	Criteria         []store.AcceptanceCriterion   `json:"criteria"`
	Relationships    []store.FeatureRelationship   `json:"relationships"`
	ChangeLinks      []store.FeatureChangeLink     `json:"change_links"`
	CapabilityStates []store.CapabilityStateRecord `json:"capability_state_history"`
	Artifacts        []store.DossierArtifact       `json:"dossier_artifacts"`
	Completed        []store.Task                  `json:"completed_tasks"`
}

type planView struct {
	store.Plan
	ItemTitle string `json:"item_title"`
}

type specification struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

type milestone struct {
	Item store.Item
	Date string
	Kind string
}

// Build writes deterministic SvelteKit source files from persisted Cassor state.
func Build(repositoryRoot, stateDir string) (string, error) {
	data, err := readData(stateDir)
	if err != nil {
		return "", err
	}
	data.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	root := filepath.Join(repositoryRoot, sourceDocsDirectory)
	if err := os.RemoveAll(root); err != nil {
		return "", fmt.Errorf("clear generated roadmap data: %w", err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", fmt.Errorf("create generated roadmap data directory: %w", err)
	}
	output := filepath.Join(root, "roadmap.json")
	if err := os.WriteFile(output, append(encoded, '\n'), 0o644); err != nil {
		return "", fmt.Errorf("write generated roadmap data: %w", err)
	}
	if err := writeGuides(repositoryRoot); err != nil {
		return "", err
	}
	return output, nil
}

// Compile builds the generated SvelteKit source into a hostable static site.
func Compile(repositoryRoot string) (string, error) {
	workspace := filepath.Join(repositoryRoot, "docs-site")
	if _, err := os.Stat(filepath.Join(workspace, "package.json")); err != nil {
		return "", fmt.Errorf("read documentation workspace: %w", err)
	}
	if err := os.RemoveAll(filepath.Join(repositoryRoot, "docs", "roadmap")); err != nil {
		return "", fmt.Errorf("clear previous documentation output: %w", err)
	}
	command := exec.Command("npm", "--prefix", workspace, "run", "build")
	if output, err := command.CombinedOutput(); err != nil {
		return "", fmt.Errorf("build documentation site: %w\n%s", err, output)
	}
	output := filepath.Join(repositoryRoot, relativeOutput)
	if _, err := os.Stat(output); err != nil {
		return "", fmt.Errorf("read generated documentation site: %w", err)
	}
	return output, nil
}

// Validate verifies that the current state can be rendered without writing files.
func Validate(stateDir string) error {
	data, err := readData(stateDir)
	if err != nil {
		return err
	}
	_, err = json.Marshal(data)
	return err
}

func writeGuides(repositoryRoot string) error {
	guidesRoot := filepath.Join(repositoryRoot, sourceGuidesDirectory)
	if err := os.MkdirAll(guidesRoot, 0o755); err != nil {
		return fmt.Errorf("create generated guide routes: %w", err)
	}
	for source, slug := range map[string]string{
		"CASSOR_CODEX_HANDOFF.md":      "project-handoff",
		"CASSOR_PLAN_PACKET_SCHEMA.md": "plan-packet-schema",
		"CASSOR_SKILLS_SPEC.md":        "skills-specification",
		"LIVING_APPLICATION_MAP.md":    "living-application-map",
		"CASSOR_RECOVERY.md":            "recovery",
		"CASSOR_CONTEXT_CONTRACT.md":    "context-contract",
		"SDD_LITE_MIGRATION_POLICY.md":  "migration-policy",
		"GETTING_STARTED.md":             "getting-started",
		"WORKFLOW_GUIDE.md":              "workflow",
		"RESUME_WORK.md":                 "resume-work",
		"ROADMAP_GUIDE.md":               "roadmap-guide",
	} {
		content, err := os.ReadFile(filepath.Join(repositoryRoot, "docs", source))
		if err != nil {
			return fmt.Errorf("read %s: %w", source, err)
		}
		guideDirectory := filepath.Join(guidesRoot, slug)
		if err := os.RemoveAll(guideDirectory); err != nil {
			return fmt.Errorf("clear generated guide route: %w", err)
		}
		output := filepath.Join(guideDirectory, "+page.md")
		if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
			return fmt.Errorf("create guide route directory: %w", err)
		}
		if err := os.WriteFile(output, append([]byte(frontmatter(strings.TrimSuffix(source, ".md"), "Repository reference documentation.")+"\n"), content...), 0o644); err != nil {
			return fmt.Errorf("write generated guide route: %w", err)
		}
	}
	return nil
}

func readData(stateDir string) (roadmapData, error) {
	projectConfig, err := config.Read(config.Path(stateDir))
	if err != nil {
		return roadmapData{}, fmt.Errorf("read configuration: %w", err)
	}
	database, err := store.Open(filepath.Join(stateDir, store.DatabaseFileName))
	if err != nil {
		return roadmapData{}, err
	}
	defer database.Close()
	return loadData(database, projectConfig.Name)
}

func loadData(database *sql.DB, projectName string) (roadmapData, error) {
	items, err := store.ListItems(database)
	if err != nil {
		return roadmapData{}, err
	}
	rows, err := database.Query(`SELECT p.id,p.roadmap_item_id,p.revision,p.content,p.status,p.active,p.created_at,p.approved_at,p.approval_note,i.title FROM plan_revisions p JOIN roadmap_items i ON i.id=p.roadmap_item_id ORDER BY p.roadmap_item_id,p.revision`)
	if err != nil {
		return roadmapData{}, err
	}
	defer rows.Close()
	plans := []planView{}
	for rows.Next() {
		var view planView
		if err := rows.Scan(&view.ID, &view.ItemID, &view.Revision, &view.Content, &view.Status, &view.Active, &view.CreatedAt, &view.ApprovedAt, &view.ApprovalNote, &view.ItemTitle); err != nil {
			return roadmapData{}, err
		}
		plans = append(plans, view)
	}
	if err := rows.Err(); err != nil {
		return roadmapData{}, err
	}
	taskRows, err := database.Query(`SELECT id,plan_revision_id,title,description,status,outcome,created_at,started_at,completed_at,blocked_at FROM tasks ORDER BY plan_revision_id,id`)
	if err != nil {
		return roadmapData{}, err
	}
	defer taskRows.Close()
	tasks := []store.Task{}
	for taskRows.Next() {
		var task store.Task
		if err := taskRows.Scan(&task.ID, &task.PlanID, &task.Title, &task.Description, &task.Status, &task.Outcome, &task.CreatedAt, &task.StartedAt, &task.CompletedAt, &task.BlockedAt); err != nil {
			return roadmapData{}, err
		}
		tasks = append(tasks, task)
	}
	if err := taskRows.Err(); err != nil {
		return roadmapData{}, err
	}
	reports, err := store.ListFeatureReports(database)
	if err != nil {
		return roadmapData{}, err
	}
	phases, err := store.ListPhaseRecords(database)
	if err != nil {
		return roadmapData{}, err
	}
	criteria, err := store.ListAcceptanceCriteria(database)
	if err != nil {
		return roadmapData{}, err
	}
	relationships, err := store.ListAllFeatureRelationships(database)
	if err != nil {
		return roadmapData{}, err
	}
	changeLinks, err := store.ListAllFeatureChangeLinks(database)
	if err != nil {
		return roadmapData{}, err
	}
	capabilityStates, err := store.ListCapabilityStateHistory(database)
	if err != nil {
		return roadmapData{}, err
	}
	artifacts, err := store.ListDossierArtifacts(database)
	if err != nil {
		return roadmapData{}, err
	}
	completedRows, err := database.Query(`SELECT id,plan_revision_id,title,description,status,outcome,created_at,started_at,completed_at,blocked_at FROM tasks WHERE status='Done' ORDER BY completed_at DESC,id DESC`)
	if err != nil {
		return roadmapData{}, err
	}
	defer completedRows.Close()
	completed := []store.Task{}
	for completedRows.Next() {
		var task store.Task
		if err := completedRows.Scan(&task.ID, &task.PlanID, &task.Title, &task.Description, &task.Status, &task.Outcome, &task.CreatedAt, &task.StartedAt, &task.CompletedAt, &task.BlockedAt); err != nil {
			return roadmapData{}, err
		}
		completed = append(completed, task)
	}
	return roadmapData{ProjectName: projectName, Items: items, Plans: plans, Tasks: tasks, Reports: reports, Phases: phases, Criteria: criteria, Relationships: relationships, ChangeLinks: changeLinks, CapabilityStates: capabilityStates, Artifacts: artifacts, Completed: completed}, completedRows.Err()
}

func renderSources(data roadmapData, repositoryRoot string) (map[string][]byte, error) {
	files := map[string][]byte{"index.md": []byte(roadmapPage(data))}
	for _, item := range data.Items {
		files[filepath.Join("roadmap", fmt.Sprintf("rm-%d.md", item.ID))] = []byte(featurePage(item))
	}
	if repositoryRoot != "" {
		for source, target := range map[string]string{
			"CASSOR_CODEX_HANDOFF.md":      "guides/project-handoff.md",
			"CASSOR_PLAN_PACKET_SCHEMA.md": "guides/plan-packet-schema.md",
			"CASSOR_SKILLS_SPEC.md":        "guides/skills-specification.md",
			"LIVING_APPLICATION_MAP.md":    "guides/living-application-map.md",
			"CASSOR_RECOVERY.md":            "guides/recovery.md",
			"CASSOR_CONTEXT_CONTRACT.md":    "guides/context-contract.md",
			"SDD_LITE_MIGRATION_POLICY.md":  "guides/migration-policy.md",
			"GETTING_STARTED.md":             "guides/getting-started.md",
			"WORKFLOW_GUIDE.md":              "guides/workflow.md",
			"RESUME_WORK.md":                 "guides/resume-work.md",
			"ROADMAP_GUIDE.md":               "guides/roadmap-guide.md",
		} {
			content, err := os.ReadFile(filepath.Join(repositoryRoot, "docs", source))
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", source, err)
			}
			files[target] = []byte(frontmatter(strings.TrimSuffix(source, ".md"), "Repository reference documentation.") + "\n" + string(content))
		}
	}
	return files, nil
}

func roadmapPage(data roadmapData) string {
	milestones := timelineMilestones(data)
	var page strings.Builder
	page.WriteString(frontmatter("Roadmap", "Approved work, delivery progress, and feature details for "+data.ProjectName+"."))
	page.WriteString(`
<style>
.timeline{position:relative;margin:2rem 0}.timeline:before{background:var(--sl-color-gray-4);content:"";left:1rem;position:absolute;top:0;bottom:0;width:2px}.milestone{display:grid;grid-template-columns:2.5rem minmax(0,1fr);gap:1rem;position:relative;margin:1.4rem 0}.dot{background:var(--sl-color-accent);border:4px solid var(--sl-color-bg);border-radius:50%;height:1.15rem;margin:.2rem 0 0 .43rem;width:1.15rem;z-index:1}.milestone.done .dot{background:#2ea043}.milestone.unscheduled .dot{background:var(--sl-color-gray-4)}.card{border:1px solid var(--sl-color-gray-5);border-radius:.65rem;color:inherit;display:block;padding:1rem;text-decoration:none}.card:hover{border-color:var(--sl-color-accent);box-shadow:0 4px 14px #0002}.date{color:var(--sl-color-accent-high);font-size:.8rem;font-weight:700;letter-spacing:.08em;text-transform:uppercase}.meta{color:var(--sl-color-gray-2);font-size:.9rem}.unscheduled{margin-top:2.5rem}@media(min-width:50rem){.timeline:before{left:50%;}.milestone{grid-template-columns:1fr 3rem 1fr}.milestone:nth-child(odd) .card{grid-column:1;text-align:right}.milestone:nth-child(odd) .dot{grid-column:2}.milestone:nth-child(odd) .card{grid-row:1}.milestone:nth-child(even) .dot{grid-column:2}.milestone:nth-child(even) .card{grid-column:3}.dot{margin:.2rem auto}}
</style>

# ` + markdownText(data.ProjectName) + ` delivery roadmap

This timeline combines verified delivery history with planned target dates. Select a milestone for its technical detail.

<section class="timeline">`)
	for _, milestone := range milestones {
		if milestone.Kind == "unscheduled" {
			page.WriteString(`</section><h2>Unscheduled</h2><section class="timeline unscheduled">`)
		}
		item := milestone.Item
		page.WriteString(`
<article class="milestone ` + milestone.Kind + `"><span class="dot"></span><a class="card" href="/roadmap/rm-` + strconv.FormatInt(item.ID, 10) + `/"><span class="date">` + milestone.Date + `</span><br><strong>RM-` + strconv.FormatInt(item.ID, 10) + ` · ` + markdownText(item.Title) + `</strong><br><span class="meta">` + markdownText(item.Status) + ` · ` + strconv.Itoa(item.Progress) + `% complete</span></a></article>`)
	}
	return page.String() + "\n</section>\n"
}

func timelineMilestones(data roadmapData) []milestone {
	plans := map[int64]int64{}
	for _, plan := range data.Plans {
		plans[plan.ID] = plan.ItemID
	}
	completed := map[int64]string{}
	for _, task := range data.Completed {
		if task.CompletedAt != nil {
			if itemID := plans[task.PlanID]; itemID != 0 && *task.CompletedAt > completed[itemID] {
				completed[itemID] = (*task.CompletedAt)[:10]
			}
		}
	}
	values := make([]milestone, 0, len(data.Items))
	for _, item := range data.Items {
		value := milestone{Item: item, Date: item.TargetDate, Kind: "planned"}
		if item.Status == "Done" {
			value.Date, value.Kind = completed[item.ID], "done"
		}
		if value.Date == "" {
			value.Date, value.Kind = "Unscheduled", "unscheduled"
		}
		values = append(values, value)
	}
	sort.SliceStable(values, func(i, j int) bool {
		if values[i].Kind == "unscheduled" {
			return false
		}
		if values[j].Kind == "unscheduled" {
			return true
		}
		return values[i].Date < values[j].Date
	})
	return values
}

func featurePage(item store.Item) string {
	var page strings.Builder
	page.WriteString(frontmatter("RM-"+strconv.FormatInt(item.ID, 10)+" · "+item.Title, valueOr(item.Description, item.Rationale)))
	page.WriteString("\n# RM-" + strconv.FormatInt(item.ID, 10) + " · " + markdownText(item.Title) + "\n\n")
	page.WriteString("**Status:** " + markdownText(item.Status) + "  \n**Progress:** " + strconv.Itoa(item.Progress) + "%  \n**Category:** " + markdownText(item.Category) + "  \n**Horizon:** " + markdownText(item.Horizon) + "\n\n")
	page.WriteString("## Summary\n\n" + markdownText(valueOr(item.TechnicalSummary, valueOr(item.Description, item.Rationale))) + "\n\n")
	page.WriteString("## Delivery details\n\n")
	for _, detail := range [][2]string{{"Target date", item.TargetDate}, {"Team", item.Team}, {"Lead engineer", item.LeadEngineer}, {"Priority", item.Priority}, {"Complexity", item.Complexity}} {
		page.WriteString("- **" + detail[0] + ":** " + markdownText(valueOr(detail[1], "Unassigned")) + "\n")
	}
	page.WriteString("\n## Technical specifications\n")
	for _, spec := range decodeSpecifications(item.Specifications) {
		marker := " "
		if spec.Done {
			marker = "x"
		}
		page.WriteString("\n- [" + marker + "] " + markdownText(spec.Title))
		if spec.Description != "" {
			page.WriteString(" — " + markdownText(spec.Description))
		}
	}
	links := decodeLinks(item.DocumentationLinks)
	if len(links) > 0 {
		page.WriteString("\n\n## Related documentation\n")
		for _, link := range links {
			page.WriteString("\n- " + documentationLink(link))
		}
	}
	return page.String() + "\n"
}

func decodeSpecifications(raw string) []specification {
	values := []specification{}
	if json.Unmarshal([]byte(raw), &values) == nil {
		return values
	}
	var labels []string
	if json.Unmarshal([]byte(raw), &labels) == nil {
		for _, label := range labels {
			values = append(values, specification{Title: label})
		}
	}
	return values
}

func decodeLinks(raw string) []string {
	var links []string
	if json.Unmarshal([]byte(raw), &links) != nil {
		return nil
	}
	return links
}

func documentationLink(path string) string {
	routes := map[string]string{
		"docs/CASSOR_CODEX_HANDOFF.md":      "/guides/project-handoff/",
		"docs/CASSOR_PLAN_PACKET_SCHEMA.md": "/guides/plan-packet-schema/",
		"docs/CASSOR_SKILLS_SPEC.md":        "/guides/skills-specification/",
	}
	if route, ok := routes[path]; ok {
		return "[" + markdownText(path) + "](" + route + ")"
	}
	return "`" + strings.ReplaceAll(path, "`", "") + "`"
}

func frontmatter(title, description string) string {
	return "---\ntitle: " + strconv.Quote(title) + "\ndescription: " + strconv.Quote(description) + "\n---\n"
}

func markdownText(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "*", "\\*", "_", "\\_", "[", "\\[", "]", "\\]", "<", "&lt;", ">", "&gt;")
	return replacer.Replace(value)
}

func valueOr(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
