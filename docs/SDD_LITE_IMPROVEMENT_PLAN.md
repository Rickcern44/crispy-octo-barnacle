# SDD Lite Improvement Plan

Status: Proposed; implementation has not been approved.
Date: 2026-09-08

## Objective

Make Cassor a token-efficient, resumable specification-driven development
workflow with a living roadmap backed by SQLite. Prioritize trustworthy workflow
state and focused context retrieval before expanding orchestration or UI features.

Retain SQLite and the CLI as the working foundation. Evolve the product direction
described in [Living Application Map](LIVING_APPLICATION_MAP.md) incrementally.

## Review baseline

The initial review examined the current working tree, including uncommitted
changes. No implementation changes were made, and `go test ./...` passed.
The findings below came from code inspection; the passing suite does not establish
that these gaps are covered.

- Approved plan criteria are stored in plan JSON, while the completion gate checks
  a separate acceptance-criteria table that plan recording does not populate.
- Packet validation permits empty acceptance criteria and task verification lists.
- Approving a new revision leaves older revisions executable because they remain
  approved.
- Compact context includes unbounded backlog and draft-plan content, but omits
  approved plans, criteria, and pending tasks needed to resume execution.
- Blocked tasks have no normal transition back into execution.
- `.cassor/` is Git-ignored, and the review found no backup/export/import command.

Relevant implementation: `internal/store/packet.go`, `internal/store/workflow.go`,
`internal/store/context.go`, and `internal/store/lifecycle.go`.

## Phase 1: Enforce the approved delivery contract

### Work

1. Record the approved plan, revision-scoped acceptance criteria, and tasks in one
   transaction. Give criteria stable identifiers and define how amendments map
   unchanged criteria into a new revision.
2. Validate meaningful criteria and task verification requirements before approval.
   Define an explicit exception policy if any work may legitimately omit them.
3. Separate historical approval from current execution authority. Track the active
   plan revision and explicitly carry forward, replace, or cancel outstanding tasks
   when an amendment becomes active.
4. Gate completion against the applicable approved criteria and required evidence.
   Record waiver attribution and reason. Preserve prior verification history.
5. Add an explicit blocked-task resume transition that retains the blocking history.
6. Extend `cassor check` to detect violations of these invariants.

### Acceptance criteria

- Recording a plan cannot leave its criteria absent from the completion model.
- A failed recording operation leaves no partial item, plan, task, or criterion data.
- Unresolved applicable criteria prevent completion.
- Superseded work cannot start under historical approval alone.
- A blocked task can resume and complete without losing its blocking reason.
- Existing databases have a documented migration policy for legacy plans and
  criteria; missing evidence is not silently converted into a pass.

### Verification

Add store and CLI tests covering atomic rollback, incomplete criteria, amendments,
stale task execution, waiver records, legacy migration, and block/resume/complete.
Run the Go suite and workflow checks against representative migrated state.

## Phase 2: Make context retrieval bounded and resumable

### Work

1. Introduce item- and role-scoped context projections. A proposed interface is
   `cassor context --item 17 --role implementation --max-bytes 12000`.
2. Include the active approved revision, relevant constraints, criteria, active or
   next task, blockers, and evidence references required by the selected role.
3. Keep default context a small navigation summary rather than a backlog dump.
4. Add explicit size limits, truncation metadata, and commands or references for
   retrieving omitted detail. Never silently omit required execution constraints.
5. Return compact mutation receipts containing identifiers, resulting state, and
   the next available action rather than routinely echoing whole artifacts.
6. Consider a transactional verification command that records evidence and
   completes the applicable task without repeated round trips.

### Acceptance criteria

- A fresh agent can identify the approved assignment, constraints, outstanding
  criteria, and next action from scoped context without reading full history.
- Increasing unrelated backlog does not increase an item's scoped context output.
- Output respects its documented bound and clearly signals incomplete context.
- Required details remain retrievable through stable identifiers.

### Verification

Use small and large fixture roadmaps to check size bounds, role relevance, and
stable output. Exercise a fresh-session resumption scenario and measure retrieval
calls and output bytes. Report token usage only when runtime telemetry provides it.

## Phase 3: Establish recovery and collaboration semantics

### Work

1. Decide whether branches and worktrees have independent state or share project
   state, including how approvals relate to repository revisions.
2. Add versioned deterministic export/import for backup, review, and transfer.
3. Define validation, conflict behavior, and transactional failure handling for
   imports. Preserve identifiers, relationships, approval history, and evidence.
4. Document the boundary between authoritative SQLite state, portable exports,
   and the generated roadmap.

### Acceptance criteria

- Export/import into a fresh database preserves the supported logical state.
- Equivalent state produces deterministic exports suitable for meaningful diffs.
- Invalid or incompatible imports fail without partially modifying live state.
- Users have a documented recovery procedure and explicit worktree behavior.

### Verification

Test round trips, incompatible versions, corrupt input, conflicting records, and
interrupted or failed imports. Verify recovery using a disposable project.

## Phase 4: Separate enduring capabilities from delivery changes

### Work

1. Refine the dossier model around enduring capabilities and linked changes.
2. Keep approved plan revisions, tasks, criteria, and evidence attached to the
   relevant delivery change.
3. Refresh accepted capability summaries after verified delivery while preserving
   the history supporting those summaries.
4. Distinguish draft findings from accepted product truth. Define when explicitly
   requested discovery may persist provisional artifacts before execution approval.
5. Expose current capabilities, in-flight changes, planned work, and known gaps as
   projections of the same underlying state.

### Acceptance criteria

- One capability can receive multiple changes without losing prior delivery history.
- Current product behavior and proposed behavior are distinguishable in both CLI
  context and roadmap views.
- Draft findings cannot silently become execution authority or accepted truth.
- A completed change has a traceable relationship to its capability update.

### Verification

Model a capability receiving two successive changes and a known gap. Verify that
context and roadmap views preserve current behavior, proposals, and prior evidence.

## Phase 5: Simplify skills and measure efficiency

### Work

1. Keep the seven lifecycle phases as guidance, with proportionate artifacts and a
   small set of enforced gates: execution readiness, amendment, and completion.
2. Move runtime-specific dispatch rules and model defaults out of the entry skill
   into references loaded only when delegation is useful.
3. Remove repeated instructions across the entry skill and protocol while keeping
   authority and approval boundaries clear.
4. Benchmark representative small fixes, features, amendments, and resumed tasks.
   Compare successful outcomes first, then calls, elapsed time, repair effort, and
   runtime-reported tokens where available.
5. Retain aggregate benchmark results outside canonical product state. Do not
   persist raw transcripts or hidden reasoning.

### Acceptance criteria

- Small changes do not require artifacts that add no decision or verification value.
- Skills direct agents through the CLI's enforced contract consistently.
- Runtime details are loaded on demand rather than for every task.
- Efficiency claims are backed by comparable runs with quality held constant;
  unavailable token telemetry remains explicitly unavailable.

## Spikes and unresolved decisions

Use [spikes/](spikes/) for bounded investigations before committing to uncertain
design choices. Suggested investigations, in priority order:

1. **Revision and criterion identity:** active authority, carry-forward behavior,
   evidence validity, and migration of legacy plan JSON.
2. **Context contract:** minimum role-specific fields, size limits, overflow
   behavior, and fresh-session resumption.
3. **State portability:** export format, import conflicts, and branch/worktree
   ownership.
4. **Capability/change model:** incremental evolution of existing dossier fields
   versus separate entities, and promotion of draft findings.
5. **Efficiency benchmark:** representative tasks, comparable starting state,
   quality criteria, and available measurement sources.

## Delivery order and scope

Deliver Phases 1–3 before expanding the application map substantially. Phase 4
builds on their authority and recovery guarantees. Prompt deduplication in Phase 5
can happen earlier, but efficiency conclusions should include scoped retrieval.

Each phase should become a bounded implementation plan with its remaining design
decisions resolved before execution. This document does not authorize code changes,
schema migrations, roadmap mutations, deployment, or a rewrite.
