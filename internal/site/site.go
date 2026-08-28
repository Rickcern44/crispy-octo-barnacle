// Package site renders Cassor's portable static roadmap projection.
package site

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/rickcern44/cassor/internal/config"
	"github.com/rickcern44/cassor/internal/store"
)

const relativeOutput = "docs/roadmap/index.html"

type roadmapData struct {
	ProjectName string       `json:"project_name"`
	Items       []store.Item `json:"items"`
	Plans       []planView   `json:"plans"`
	Completed   []store.Task `json:"completed_tasks"`
}
type planView struct {
	store.Plan
	ItemTitle string `json:"item_title"`
}

// Build writes the deterministic, standalone roadmap page below repositoryRoot.
func Build(repositoryRoot, stateDir string) (string, error) {
	projectConfig, err := config.Read(config.Path(stateDir))
	if err != nil {
		return "", fmt.Errorf("read configuration: %w", err)
	}
	database, err := store.Open(filepath.Join(stateDir, store.DatabaseFileName))
	if err != nil {
		return "", err
	}
	defer database.Close()
	data, err := loadData(database, projectConfig.Name)
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("encode roadmap data: %w", err)
	}
	page, err := render(template.JS(encoded))
	if err != nil {
		return "", err
	}
	output := filepath.Join(repositoryRoot, relativeOutput)
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return "", fmt.Errorf("create roadmap directory: %w", err)
	}
	if err := os.WriteFile(output, page, 0o644); err != nil {
		return "", fmt.Errorf("write roadmap: %w", err)
	}
	return output, nil
}

// Validate verifies that the current state can be serialized and rendered
// without changing the generated roadmap projection.
func Validate(stateDir string) error {
	projectConfig, err := config.Read(config.Path(stateDir))
	if err != nil {
		return fmt.Errorf("read configuration: %w", err)
	}
	database, err := store.Open(filepath.Join(stateDir, store.DatabaseFileName))
	if err != nil {
		return err
	}
	defer database.Close()
	data, err := loadData(database, projectConfig.Name)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("encode roadmap data: %w", err)
	}
	_, err = render(template.JS(encoded))
	return err
}

func loadData(database *sql.DB, projectName string) (roadmapData, error) {
	items, err := store.ListItems(database)
	if err != nil {
		return roadmapData{}, err
	}
	rows, err := database.Query(`SELECT p.id,p.roadmap_item_id,p.revision,p.content,p.status,p.created_at,p.approved_at,p.approval_note,i.title FROM plan_revisions p JOIN roadmap_items i ON i.id=p.roadmap_item_id ORDER BY p.roadmap_item_id,p.revision`)
	if err != nil {
		return roadmapData{}, err
	}
	defer rows.Close()
	plans := []planView{}
	for rows.Next() {
		var view planView
		if err := rows.Scan(&view.ID, &view.ItemID, &view.Revision, &view.Content, &view.Status, &view.CreatedAt, &view.ApprovedAt, &view.ApprovalNote, &view.ItemTitle); err != nil {
			return roadmapData{}, err
		}
		plans = append(plans, view)
	}
	if err := rows.Err(); err != nil {
		return roadmapData{}, err
	}
	completedRows, err := database.Query(`SELECT id,plan_revision_id,title,description,status,outcome,created_at,started_at,completed_at,blocked_at FROM tasks WHERE status='Completed' ORDER BY completed_at DESC,id DESC`)
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
	return roadmapData{ProjectName: projectName, Items: items, Plans: plans, Completed: completed}, completedRows.Err()
}

func render(data template.JS) ([]byte, error) {
	var output bytes.Buffer
	if err := pageTemplate.Execute(&output, struct{ Data template.JS }{data}); err != nil {
		return nil, fmt.Errorf("render roadmap: %w", err)
	}
	return output.Bytes(), nil
}

var pageTemplate = template.Must(template.New("roadmap").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Cassor roadmap</title><style>
:root{color-scheme:light dark;--bg:#f7f8fb;--card:#fff;--ink:#172033;--muted:#5e6b82;--line:#dbe1ec;--accent:#635bff}.dark{--bg:#111521;--card:#1a2030;--ink:#f4f6fb;--muted:#aab5cb;--line:#303a50;--accent:#a9a5ff}*{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--ink);font:15px system-ui,sans-serif}main{max-width:1100px;margin:auto;padding:32px 20px}header{display:flex;justify-content:space-between;gap:18px;align-items:start}h1{margin:0;font-size:2rem}h2{margin:32px 0 12px}p{color:var(--muted)}button,select{font:inherit;padding:8px 10px;border:1px solid var(--line);border-radius:7px;background:var(--card);color:var(--ink)}.filters{display:flex;flex-wrap:wrap;gap:8px;margin:24px 0}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(250px,1fr));gap:12px}.card{background:var(--card);border:1px solid var(--line);border-radius:10px;padding:16px}.meta{color:var(--muted);font-size:.86rem}.tag{display:inline-block;background:color-mix(in srgb,var(--accent) 15%,transparent);color:var(--accent);border-radius:999px;padding:3px 8px;margin:8px 5px 0 0;font-size:.8rem}.empty{color:var(--muted);font-style:italic}@media(max-width:600px){main{padding:22px 14px}header{display:block}header button{margin-top:14px}}
</style></head><body><main><header><div><p class="meta">CASSOR ROADMAP</p><h1 id="project"></h1><p>Approved work, proposals, and verified delivery history.</p></div><button id="theme">Toggle theme</button></header><div class="filters"><select id="category"><option value="">All categories</option></select><select id="horizon"><option value="">All horizons</option></select><select id="status"><option value="">All statuses</option></select></div><h2>Roadmap</h2><div class="grid" id="items"></div><h2>Plan approval state</h2><div class="grid" id="plans"></div><h2>Completed-task history</h2><div class="grid" id="completed"></div></main><script>const data={{.Data}};const $=id=>document.getElementById(id),esc=s=>String(s??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));const card=(title,body)=>'<article class="card"><strong>'+esc(title)+'</strong>'+body+'</article>';$('project').textContent=data.project_name;for(const [id,key]of [['category','category'],['horizon','horizon'],['status','status']]){[...new Set(data.items.map(x=>x[key]))].sort().forEach(v=>$(id).insertAdjacentHTML('beforeend','<option>'+esc(v)+'</option>'))}function render(){const visible=data.items.filter(x=>(!$('category').value||x.category===$('category').value)&&(!$('horizon').value||x.horizon===$('horizon').value)&&(!$('status').value||x.status===$('status').value));$('items').innerHTML=visible.length?visible.map(x=>card(x.title,'<div class="meta">'+esc(x.description||x.rationale||'No description')+'</div><span class="tag">'+esc(x.category)+'</span><span class="tag">'+esc(x.horizon)+'</span><span class="tag">'+esc(x.status)+'</span>')).join(''):'<p class="empty">No roadmap items match these filters.</p>';$('plans').innerHTML=data.plans.length?data.plans.map(x=>card(x.item_title,'<div class="meta">Revision '+x.revision+'</div><span class="tag">'+esc(x.status)+'</span>')).join(''):'<p class="empty">No plans yet.</p>';$('completed').innerHTML=data.completed_tasks.length?data.completed_tasks.map(x=>card(x.title,'<div class="meta">'+esc(x.outcome||'Completed')+'</div>')).join(''):'<p class="empty">No completed tasks yet.</p>'}document.querySelectorAll('select').forEach(x=>x.onchange=render);$('theme').onclick=()=>document.documentElement.classList.toggle('dark');render();</script></body></html>`))
