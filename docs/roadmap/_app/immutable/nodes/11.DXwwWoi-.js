import{C as e,H as t,T as n,W as r,nt as i,rt as a,v as o}from"../chunks/CubxcVna.js";import"../chunks/xihTtKlq.js";import"../chunks/CamEucdS.js";var{title:s,description:c}={title:`CASSOR_RECOVERY`,description:`Repository reference documentation.`},l=n(`<h1>Cassor recovery and checkout ownership</h1> <p>SQLite in <code>.cassor/cassor.db</code> is the authoritative state for one checkout.
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
checkouts.</p>`,1);function u(n){var s=l(),c=r(t(s),6);o(c,()=>`<code class="language-text">cassor export --file /secure/path/cassor-state.json</code>`,!0),a(c);var u=r(c,4);o(u,()=>`<code class="language-text">cassor init --name &quot;Cassor project&quot;
cassor migrate
cassor import --file /secure/path/cassor-state.json
cassor check
cassor site build</code>`,!0),a(u),i(4),e(n,s)}export{u as component};