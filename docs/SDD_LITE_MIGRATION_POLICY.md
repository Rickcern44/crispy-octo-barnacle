# SDD Lite delivery-contract migration policy

This historical-schema operational reference is subordinate to the current
[Cassor PRD](PRD.md).

Migration `0011_delivery_contract` adds active plan authority, revision-scoped
criteria, task verification requirements, task event history, and criterion
evidence history.

For an existing roadmap item, the latest approved plan revision becomes active;
older approved revisions remain historical. Packet-shaped plan content is read
to recover its criteria and task verification commands. A free-form legacy plan
gets a pending criterion named `legacy-plan-<plan-id>`, and a legacy task keeps
an empty verification list. Neither case is treated as verified. The generated
roadmap and `cassor check` therefore expose missing contract data on executable
legacy work, while historical completed items retain their recorded status.

New plans must contain at least one meaningful criterion and task verification
requirement before approval. A criterion is complete only with evidence; a
waiver additionally requires the approver identity and a reason. Prior
verification attempts remain in `criterion_evidence`, and blocking/resume
transitions remain in `task_events`.

To recover an old database, copy the `.cassor` directory, run `cassor migrate`,
review `cassor check` and the pending criteria, then verify or explicitly waive
each applicable legacy criterion before completing its roadmap item. Migration
is transactional; if the backfill fails, the schema migration and its data
changes are rolled back together.
