import{S as e,U as t,V as n,_ as r,nt as i,tt as a,w as o}from"../chunks/DWbyJt3s.js";import"../chunks/xihTtKlq.js";import"../chunks/Bnb0ZNYR.js";var{title:s,description:c}={title:`CASSOR_CONTEXT_CONTRACT`,description:`Repository reference documentation.`},l=o(`<h1>Cassor context contract</h1> <p><code>cassor context</code> returns a bounded navigation summary with counts and stable
commands for the first ten entries in each queue. It does not include plan
content, task descriptions, or the backlog payload.</p> <p>Use scoped context to resume an item:</p> <pre class="language-text"></pre> <p>Supported roles are <code>implementation</code>, <code>verification</code>, and <code>planning</code>. The
response identifies the item, active approved plan revision, packet goal and
constraints, revision-scoped criteria, the next task and its verification
requirements, blockers, and evidence references. Every omitted detail includes
an ID and a command for retrieving it with <code>cassor plan show</code>, <code>cassor task show</code>, or the lifecycle criterion command.</p> <p><code>max-bytes</code> measures compact JSON bytes. Required fields remain present. If
optional detail cannot fit, <code>truncated</code> is true and <code>omitted_details</code> lists
stable references. If the bound is smaller than the required projection, the
command fails with the minimum size instead of returning incomplete execution
constraints. Mutation commands return IDs, resulting state, and a next action
in their default text output; pass <code>--json</code> where the full recorded artifact is
needed explicitly.</p>`,1);function u(o){var s=l(),c=t(n(s),6);r(c,()=>`<code class="language-text">cassor context --item 17 --role implementation --max-bytes 12000 --json</code>`,!0),i(c),a(4),e(o,s)}export{u as component};