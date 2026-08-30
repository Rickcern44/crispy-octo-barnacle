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
	return roadmapData{ProjectName: projectName, Items: items, Plans: plans, Completed: completed}, completedRows.Err()
}

func render(data template.JS) ([]byte, error) {
	var output bytes.Buffer
	if err := carbonTemplate.Execute(&output, struct{ Data template.JS }{data}); err != nil {
		return nil, fmt.Errorf("render roadmap: %w", err)
	}
	return output.Bytes(), nil
}

var carbonTemplate = template.Must(template.New("carbon").Parse(`<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Cassor · Developer Roadmap</title><style>:root{--bg:#161616;--panel:#262626;--line:#393939;--ink:#f4f4f4;--muted:#a8a8a8;--blue:#0f62fe;--green:#42be65;--red:#fa4d56}*{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--ink);font:14px 'IBM Plex Sans',Arial,sans-serif}main{max-width:1440px;margin:auto;padding:28px}header{border-bottom:1px solid var(--line);padding-bottom:20px;display:flex;justify-content:space-between}h1{margin:4px 0;font-size:30px}.eyebrow{color:#78a9ff;font-size:12px;font-weight:bold;letter-spacing:.1em}.stats,.triage{display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr));gap:1px;background:var(--line);margin:20px 0}.stat,.lane{background:var(--panel);padding:14px}.stat b{font-size:28px;color:#78a9ff;display:block}input,select{background:#262626;border:1px solid #525252;color:white;padding:10px;margin-right:8px}.table{width:100%;border-collapse:collapse;margin-top:18px}.table th{text-align:left;background:#393939;padding:10px;color:#c6c6c6;font-size:12px}.table td{padding:12px 10px;border-bottom:1px solid var(--line)}tr{cursor:pointer}tr:hover{background:#222}.tag{padding:3px 8px;background:#393939;border-left:3px solid var(--blue);font-size:12px}.bar{height:6px;background:#393939;width:110px}.bar i{display:block;height:100%;background:var(--green)}#detail{position:fixed;right:0;top:0;height:100%;width:min(480px,100%);background:#262626;border-left:1px solid #525252;padding:24px;overflow:auto;display:none}#detail.open{display:block}.spec{padding:9px 0;border-bottom:1px solid #393939;color:#c6c6c6}@media(max-width:700px){main{padding:16px}.table th:nth-child(4),.table td:nth-child(4),.table th:nth-child(5),.table td:nth-child(5){display:none}}</style></head><body><main><header><div><div class="eyebrow">CASSOR / DEVELOPER ROADMAP</div><h1 id="name"></h1><div style="color:#a8a8a8">Execution visibility for technical delivery.</div></div></header><section class="stats" id="stats"></section><div><input id="q" placeholder="Filter by ID or feature"><select id="status"><option value="">All statuses</option></select></div><table class="table"><thead><tr><th>ID / FEATURE</th><th>STATUS</th><th>PROGRESS</th><th>TARGET</th><th>OWNER</th><th>PRIORITY</th></tr></thead><tbody id="rows"></tbody></table><h2>Lifecycle triage</h2><section class="triage" id="triage"></section></main><aside id="detail"></aside><script>const data={{.Data}},$=x=>document.getElementById(x),e=x=>String(x??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));const statuses=['Planned','Ready','In Progress','Blocked','Done','Won’t Do'];$('name').textContent=data.project_name;statuses.forEach(s=>$('status').insertAdjacentHTML('beforeend','<option>'+s+'</option>'));function items(){let q=$('q').value.toLowerCase(),s=$('status').value;return data.items.filter(x=>(!s||x.status===s)&&(!q||(x.title+x.id).toLowerCase().includes(q)))}function detail(x){let specs=JSON.parse(x.specifications||'[]'),docs=JSON.parse(x.documentation_links||'[]');$('detail').className='open';$('detail').innerHTML='<button onclick="this.parentElement.className=\'\'">Close</button><div class="eyebrow">RM-'+x.id+'</div><h2>'+e(x.title)+'</h2><p>'+e(x.technical_summary||x.description)+'</p><p><b>Team:</b> '+e(x.team||'Unassigned')+'<br><b>Lead:</b> '+e(x.lead_engineer||'Unassigned')+'<br><b>Complexity:</b> '+e(x.complexity)+'</p><h3>Technical specifications</h3>'+specs.map(s=>'<div class="spec">□ '+e(s)+'</div>').join('')+'<h3>Documentation</h3>'+docs.map(d=>'<div class="spec">'+e(d)+'</div>').join('')}function render(){let a=items();$('stats').innerHTML=statuses.map(s=>'<div class="stat"><b>'+a.filter(x=>x.status===s).length+'</b>'+e(s)+'</div>').join('');$('rows').innerHTML=a.map(x=>'<tr onclick="detail(data.items.find(i=>i.id=='+x.id+'))"><td><b>RM-'+x.id+'</b><br>'+e(x.title)+'</td><td><span class="tag">'+e(x.status)+'</span></td><td><div class="bar"><i style="width:'+x.progress+'%"></i></div> '+x.progress+'%</td><td>'+e(x.target_date||'—')+'</td><td>'+e(x.lead_engineer||x.team||'—')+'</td><td>'+e(x.priority)+'</td></tr>').join('');let map=[['Proposed','Planned'],['Approved','Ready'],['Declined','Won’t Do']];$('triage').innerHTML=map.map(([label,s])=>'<div class="lane"><b>'+label+'</b><p>'+a.filter(x=>x.status===s).map(x=>e(x.title)).join('<br>')||'—'+'</p></div>').join('')}$('q').oninput=render;$('status').onchange=render;render()</script></body></html>`))

var dashboardTemplate = template.Must(template.New("dashboard").Parse(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Cassor dashboard</title><style>
:root{--bg:#09111f;--panel:#111c30;--card:#17253d;--ink:#eff6ff;--muted:#9bb0ca;--line:#29415f;--accent:#53d3a6;--warn:#ffc857;--danger:#ff7272}*{box-sizing:border-box}body{margin:0;background:radial-gradient(circle at top right,#173560,var(--bg) 45%);color:var(--ink);font:15px ui-sans-serif,system-ui}main{max-width:1280px;margin:auto;padding:32px 20px}header{display:flex;justify-content:space-between;align-items:end;gap:20px}.eyebrow{color:var(--accent);font-weight:700;letter-spacing:.12em;font-size:.75rem}h1{font-size:clamp(2rem,5vw,3.5rem);margin:.2rem 0}p{color:var(--muted)}button,input,select{background:var(--panel);border:1px solid var(--line);color:var(--ink);padding:10px 12px;border-radius:9px;font:inherit}.controls{display:flex;gap:8px;flex-wrap:wrap;margin:28px 0}.stats{display:grid;grid-template-columns:repeat(auto-fit,minmax(130px,1fr));gap:10px}.stat,.card{background:linear-gradient(145deg,var(--card),var(--panel));border:1px solid var(--line);border-radius:14px;padding:16px}.stat b{display:block;font-size:1.8rem;color:var(--accent)}.board{display:grid;grid-template-columns:repeat(auto-fit,minmax(240px,1fr));gap:12px;margin-top:18px}.lane h2{font-size:1rem;margin:0 0 10px}.card{margin-bottom:10px;cursor:pointer}.card:hover{border-color:var(--accent);transform:translateY(-2px)}.meta{font-size:.84rem;color:var(--muted);margin-top:8px}.tag{display:inline-block;padding:3px 8px;border-radius:999px;background:#203756;color:#cbe3ff;font-size:.78rem;margin:9px 5px 0 0}.detail{display:none;margin-top:12px;color:var(--muted)}.open .detail{display:block}.empty{color:var(--muted)}@media(max-width:650px){main{padding:20px 14px}header{display:block}.controls>*{width:100%}}
</style></head><body><main><header><div><div class="eyebrow">CASSOR · DELIVERY DASHBOARD</div><h1 id="name"></h1><p>What is planned, ready, underway, and done.</p></div><button id="theme">Contrast</button></header><section class="stats" id="stats"></section><div class="controls"><input id="search" placeholder="Search roadmap"><select id="category"><option value="">All categories</option></select><select id="horizon"><option value="">All horizons</option></select></div><section class="board" id="board"></section></main><script>const data={{.Data}},$=x=>document.getElementById(x),esc=x=>String(x??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));const statuses=['Planned','Ready','In Progress','Blocked','Done','Won’t Do'];$('name').textContent=data.project_name;for(const [id,key] of [['category','category'],['horizon','horizon']])[...new Set(data.items.map(x=>x[key]))].sort().forEach(v=>$(id).insertAdjacentHTML('beforeend','<option>'+esc(v)+'</option>'));function render(){let q=$('search').value.toLowerCase(),c=$('category').value,h=$('horizon').value,items=data.items.filter(x=>(!q||(x.title+x.description+x.rationale).toLowerCase().includes(q))&&(!c||x.category===c)&&(!h||x.horizon===h));$('stats').innerHTML=statuses.map(s=>'<div class="stat"><b>'+items.filter(x=>x.status===s).length+'</b>'+esc(s)+'</div>').join('');$('board').innerHTML=statuses.map(s=>{let cards=items.filter(x=>x.status===s).map(x=>'<article class="card" onclick="this.classList.toggle(\'open\')"><strong>'+esc(x.title)+'</strong><div class="meta">'+esc(x.category)+' · '+esc(x.horizon)+'</div><div class="detail">'+esc(x.description||x.rationale||'No additional detail.')+'</div></article>').join('')||'<p class="empty">Nothing here</p>';return '<section class="lane"><h2>'+esc(s)+'</h2>'+cards+'</section>'}).join('')}$('search').oninput=render;$('category').onchange=render;$('horizon').onchange=render;$('theme').onclick=()=>document.body.classList.toggle('contrast');render()</script></body></html>`))

var pageTemplate = template.Must(template.New("roadmap").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Cassor roadmap</title><style>
:root{color-scheme:light dark;--bg:#f7f8fb;--card:#fff;--ink:#172033;--muted:#5e6b82;--line:#dbe1ec;--accent:#635bff}.dark{--bg:#111521;--card:#1a2030;--ink:#f4f6fb;--muted:#aab5cb;--line:#303a50;--accent:#a9a5ff}*{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--ink);font:15px system-ui,sans-serif}main{max-width:1100px;margin:auto;padding:32px 20px}header{display:flex;justify-content:space-between;gap:18px;align-items:start}h1{margin:0;font-size:2rem}h2{margin:32px 0 12px}p{color:var(--muted)}button,select{font:inherit;padding:8px 10px;border:1px solid var(--line);border-radius:7px;background:var(--card);color:var(--ink)}.filters{display:flex;flex-wrap:wrap;gap:8px;margin:24px 0}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(250px,1fr));gap:12px}.card{background:var(--card);border:1px solid var(--line);border-radius:10px;padding:16px}.meta{color:var(--muted);font-size:.86rem}.tag{display:inline-block;background:color-mix(in srgb,var(--accent) 15%,transparent);color:var(--accent);border-radius:999px;padding:3px 8px;margin:8px 5px 0 0;font-size:.8rem}.empty{color:var(--muted);font-style:italic}@media(max-width:600px){main{padding:22px 14px}header{display:block}header button{margin-top:14px}}
</style></head><body><main><header><div><p class="meta">CASSOR ROADMAP</p><h1 id="project"></h1><p>Approved work, proposals, and verified delivery history.</p></div><button id="theme">Toggle theme</button></header><div class="filters"><select id="category"><option value="">All categories</option></select><select id="horizon"><option value="">All horizons</option></select><select id="status"><option value="">All statuses</option></select></div><h2>Roadmap</h2><div class="grid" id="items"></div><h2>Plan approval state</h2><div class="grid" id="plans"></div><h2>Completed-task history</h2><div class="grid" id="completed"></div></main><script>const data={{.Data}};const $=id=>document.getElementById(id),esc=s=>String(s??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));const card=(title,body)=>'<article class="card"><strong>'+esc(title)+'</strong>'+body+'</article>';$('project').textContent=data.project_name;for(const [id,key]of [['category','category'],['horizon','horizon'],['status','status']]){[...new Set(data.items.map(x=>x[key]))].sort().forEach(v=>$(id).insertAdjacentHTML('beforeend','<option>'+esc(v)+'</option>'))}function render(){const visible=data.items.filter(x=>(!$('category').value||x.category===$('category').value)&&(!$('horizon').value||x.horizon===$('horizon').value)&&(!$('status').value||x.status===$('status').value));$('items').innerHTML=visible.length?visible.map(x=>card(x.title,'<div class="meta">'+esc(x.description||x.rationale||'No description')+'</div><span class="tag">'+esc(x.category)+'</span><span class="tag">'+esc(x.horizon)+'</span><span class="tag">'+esc(x.status)+'</span>')).join(''):'<p class="empty">No roadmap items match these filters.</p>';$('plans').innerHTML=data.plans.length?data.plans.map(x=>card(x.item_title,'<div class="meta">Revision '+x.revision+'</div><span class="tag">'+esc(x.status)+'</span>')).join(''):'<p class="empty">No plans yet.</p>';$('completed').innerHTML=data.completed_tasks.length?data.completed_tasks.map(x=>card(x.title,'<div class="meta">'+esc(x.outcome||'Completed')+'</div>')).join(''):'<p class="empty">No completed tasks yet.</p>'}document.querySelectorAll('select').forEach(x=>x.onchange=render);$('theme').onclick=()=>document.documentElement.classList.toggle('dark');render();</script></body></html>`))
