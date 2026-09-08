import{S as e,tt as t,w as n}from"../chunks/DWbyJt3s.js";import"../chunks/xihTtKlq.js";import"../chunks/Bnb0ZNYR.js";var{title:r,description:i}={title:`SDD_LITE_MIGRATION_POLICY`,description:`Repository reference documentation.`},a=n(`<h1>SDD Lite delivery-contract migration policy</h1> <p>Migration <code>0011_delivery_contract</code> adds active plan authority, revision-scoped
criteria, task verification requirements, task event history, and criterion
evidence history.</p> <p>For an existing roadmap item, the latest approved plan revision becomes active;
older approved revisions remain historical. Packet-shaped plan content is read
to recover its criteria and task verification commands. A free-form legacy plan
gets a pending criterion named <code>legacy-plan-&lt;plan-id&gt;</code>, and a legacy task keeps
an empty verification list. Neither case is treated as verified. The generated
roadmap and <code>cassor check</code> therefore expose missing contract data on executable
legacy work, while historical completed items retain their recorded status.</p> <p>New plans must contain at least one meaningful criterion and task verification
requirement before approval. A criterion is complete only with evidence; a
waiver additionally requires the approver identity and a reason. Prior
verification attempts remain in <code>criterion_evidence</code>, and blocking/resume
transitions remain in <code>task_events</code>.</p> <p>To recover an old database, copy the <code>.cassor</code> directory, run <code>cassor migrate</code>,
review <code>cassor check</code> and the pending criteria, then verify or explicitly waive
each applicable legacy criterion before completing its roadmap item. Migration
is transactional; if the backfill fails, the schema migration and its data
changes are rolled back together.</p>`,1);function o(n){var r=a();t(8),e(n,r)}export{o as component};