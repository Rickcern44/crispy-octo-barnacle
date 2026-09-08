import{S as e,U as t,V as n,_ as r,nt as i,tt as a,w as o}from"../chunks/DWbyJt3s.js";import"../chunks/xihTtKlq.js";import"../chunks/Bnb0ZNYR.js";var{title:s,description:c}={title:`CASSOR_RECOVERY`,description:`Repository reference documentation.`},l=o(`<h1>Cassor recovery and checkout ownership</h1> <p>SQLite in <code>.cassor/cassor.db</code> is the authoritative state for one checkout.
Generated files under <code>docs/roadmap</code> are a view and can be rebuilt with <code>cassor site build</code>. A branch or worktree does not automatically share state
with another checkout; initialize and migrate each checkout that should have a
local Cassor database.</p> <p>Create a deterministic backup from the repository root:</p> <pre class="language-text"></pre> <p>To recover, create or use a fresh initialized checkout, apply migrations, then
import the state:</p> <pre class="language-text"></pre> <p>The version 1 export includes roadmap items, categories, plan revisions and
approval state, tasks and verification requirements, lifecycle records,
criteria, criterion evidence history, task events, feature relationships, and
feature reports. It excludes schema metadata, generated roadmap files, local
configuration, transcripts, and hidden reasoning.</p> <p>Imports reject unknown format versions, malformed JSON, missing references, and
non-empty logical destinations. The category rows created by <code>cassor init</code> may
remain when their IDs and names match the export. All other records are
inserted in one transaction; a validation or constraint failure rolls the
destination back without a partial restore. Version 1 deliberately has no
merge behavior, so export first when transferring state between populated
checkouts.</p>`,1);function u(o){var s=l(),c=t(n(s),6);r(c,()=>`<code class="language-text">cassor export --file /secure/path/cassor-state.json</code>`,!0),i(c);var u=t(c,4);r(u,()=>`<code class="language-text">cassor init --name &quot;Cassor project&quot;
cassor migrate
cassor import --file /secure/path/cassor-state.json
cassor check
cassor site build</code>`,!0),i(u),a(4),e(o,s)}export{u as component};