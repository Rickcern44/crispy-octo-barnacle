import{S as e,U as t,V as n,_ as r,nt as i,tt as a,w as o}from"../chunks/DWbyJt3s.js";import"../chunks/xihTtKlq.js";import"../chunks/Bnb0ZNYR.js";var{title:s,description:c}={title:`LIVING_APPLICATION_MAP`,description:`Repository reference documentation.`},l=o(`<h1>Living Application Map</h1> <h2>Product direction</h2> <p>Cassor should grow into a living product memory for an application: a
one-stop workspace for understanding existing capabilities, defining the next
ones, approving delivery work, and retaining the evidence behind the result.</p> <p>It is not only a task tracker or an agent orchestrator.  The primary user
experience is feature authoring and application understanding; agents are
contributors to that shared, curated record.</p> <h2>Two kinds of truth</h2> <p>Each feature record should make two related things clear:</p> <table><thead><tr><th>Dimension</th><th>Question answered</th></tr></thead><tbody><tr><td>Current product truth</td><td>What does the application do today? What is supported, limited, deprecated, or known to be missing?</td></tr><tr><td>Intended change</td><td>What should change next, why, what was decided, and how will success be verified?</td></tr></tbody></table> <p>A shipped capability remains in Cassor after delivery.  For example,
Authentication retains its current behavior, constraints, decisions, and
known gaps.  A future change such as Passkeys is linked to Authentication as a
child feature or change proposal rather than replacing its history.</p> <h2>Core model</h2> <p>The enduring unit is a <strong>feature dossier</strong>, not a one-time roadmap task.</p> <pre class="language-text"></pre> <p>Existing Cassor roadmap items, plan revisions, tasks, lifecycle phase records,
acceptance criteria, and feature reports can become parts of this dossier. The
goal is an incremental evolution rather than a rewrite.</p> <h2>Agent collaboration</h2> <p>Agents contribute compact, evidence-based artifacts to a feature dossier.
For example, a research agent may submit a research finding; an implementation
agent may add an implementation note; and a verification agent may add
verification evidence.</p> <p>Each artifact should identify its kind, author role, status, summary,
supporting evidence, timestamps, and any superseded artifact.  Artifacts are
append-only or revisioned, rather than silently overwritten.</p> <p>The orchestrator remains the editorial and workflow authority:</p> <ul><li>Workers may write only artifact types within their assigned scope.</li> <li>Submitted artifacts may be visible as unreviewed, but only accepted artifacts
may influence a plan or workflow transition.</li> <li>Only the orchestrator may approve plans, waive criteria, change canonical
feature state, or make roadmap transitions.</li> <li>Persist compact conclusions and evidence, never raw worker transcripts or
hidden reasoning.</li></ul> <p>An agent’s <code>role</code> (for example, <code>research</code>, <code>implementation</code>, or <code>verification</code>) describes who contributed an artifact. It must not grant that
agent ownership of the feature’s shared current truth.</p> <h2>Context as a projection</h2> <p>Agent context should be a role- and feature-scoped view of the dossier, not a
copy of the entire database and not an opaque agent memory store.  For example,
a research context can contain the current-state summary, accepted decisions,
open questions, relevant prior findings, and known risks.  An implementation
context adds the approved plan and acceptance criteria.</p> <p>This keeps handoffs resumable while preventing stale, speculative, or
irrelevant notes from being treated as canonical.</p> <h2>Delivery sequence</h2> <p>This direction is an initiative composed of independently useful features:</p> <ol><li><strong>Feature dossier foundation</strong> — establish durable feature records with a
current-state summary, feature classification/status, and relationships.</li> <li><strong>Application map</strong> — show shipped capabilities, in-flight work, planned
work, and known gaps as complementary views.</li> <li><strong>Feature-writing workflow</strong> — guide authors through problem, intent,
scope, assumptions, decisions, and acceptance criteria.</li> <li><strong>Agent contribution artifacts</strong> — persist scoped, attributed findings with
explicit acceptance by the orchestrator.</li> <li><strong>Change proposals and delivery history</strong> — allow an enduring feature to
receive multiple approved changes over time.</li> <li><strong>Cross-feature decisions and dependencies</strong> — expose the decisions and
relationships that affect more than one feature.</li></ol> <p>The recommended first delivery is <strong>Feature dossier foundation</strong>. It creates
the durable capability record that makes the application map, guided authoring,
and agent contributions coherent.</p> <h2>Capability and change boundary</h2> <p>Cassor keeps <code>Capability</code>, <code>Change</code>, and <code>Gap</code> as distinct feature types. A
delivery change may be linked explicitly to an enduring capability with a
typed change-of dossier link. A capability can therefore receive multiple
changes without replacing its prior delivery history.</p> <p>Current product truth is stored on the capability and refreshed only through
an explicitly accepted state update sourced from a completed linked change.
Those updates are append-only, so prior accepted states remain traceable.
Discovery and other agents may record attributed dossier artifacts as <code>Draft</code>,
but drafts cannot approve plans, complete work, or become canonical product
truth until the orchestrator explicitly accepts them with evidence.</p> <p>The CLI and generated roadmap expose these as complementary projections:
capabilities describe current truth, changes describe intended or delivered
work, and gaps describe known missing behavior. SQLite remains authoritative;
the context JSON and static site are derived views.</p> <h2>Intended lifecycle</h2> <pre class="language-text"></pre> <p>Refreshing current truth after delivery is essential: Cassor should describe
the application as it is now, while preserving the decisions and evidence that
explain how it arrived there.</p>`,1);function u(o){var s=l(),c=t(n(s),20);r(c,()=>`<code class="language-text">Application
  |- shipped capabilities
  |- in-flight features
  |- planned opportunities
  |- known gaps and technical debt
  &#96;- cross-cutting decisions and constraints

Feature dossier
  |- current-state summary
  |- problem, intent, and scope
  |- research, risks, assumptions, and decisions
  |- acceptance criteria
  |- change proposals and approved delivery plans
  |- tasks, implementation record, and verification evidence
  &#96;- relationships to other features and decisions</code>`,!0),i(c);var u=t(c,40);r(u,()=>`<code class="language-text">Discover existing capability
  -&gt; document current state
  -&gt; identify a gap or desired change
  -&gt; research and define it
  -&gt; approve a delivery plan
  -&gt; implement and verify
  -&gt; refresh the feature&#39;s current-state summary</code>`,!0),i(u),a(2),e(o,s)}export{u as component};