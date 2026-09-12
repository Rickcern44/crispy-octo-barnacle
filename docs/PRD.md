# Cassor Product Requirements Document

Status: Canonical product document  
Last reviewed: 2026-09-09

## Product definition

Cassor is a **skill-first workflow for agent-assisted software development**.
It helps a person turn a request into small, approved, verifiable work while
keeping only the durable context needed to resume that work later.

The Cassor skill is the primary product experience. The repository-local CLI
and SQLite database are supporting infrastructure: they enforce workflow
boundaries, retain compact state, and produce useful projections without
becoming a product-management system of their own.

## Problem

Specification-driven development is valuable when it makes decisions and
delivery safer. It becomes counterproductive when it produces long documents
that people do not read, repeats chat context for every agent, or treats every
small task as a formal program.

Agent-assisted work also needs durable coordination. Without it, agents can
lose decisions between sessions, act on stale context, overlap work, or blur
the line between a proposed plan and approved execution.

Cassor solves those problems with a concise, proportionate workflow rather
than a document-heavy process.

## Intended user and job

The primary user is a developer or technical product owner working with an AI
coding runtime. They need to:

- understand a request enough to make a good decision;
- approve a compact execution plan before meaningful repository changes begin;
- let agents investigate, implement, and verify bounded work;
- resume interrupted work without reconstructing long conversations; and
- retain a short, trustworthy record of what changed and why.

Cassor is designed for a repository-local, single decision-maker workflow. It
does not currently target organization-wide project management, multi-writer
collaboration, cloud synchronization, or autonomous delivery without human
approval.

## Product principles

### Skill first

Users invoke one Cassor skill. Detailed role guidance is loaded only when it is
useful. The CLI is the state kernel behind that experience, not a second
workflow users must learn independently.

### Proportionate process

Small, local changes should require only the discovery and verification needed
to make them safe. Larger changes may use focused investigation, a revisioned
plan, and several sequential tasks. Process expands with uncertainty and risk,
not by default.

### Human approval at the execution boundary

Cassor distinguishes exploration from execution. A concrete plan revision is
approved before repository mutation; material changes to that approved outcome
require a new revision and approval.

### Token efficiency by construction

Token efficiency is a product-quality requirement, not simply a preference for
cheaper models. Cassor should minimize irrelevant context, repeated
instructions, full transcript storage, and unnecessary agent handoffs. It
should load focused references on demand, provide role- and work-scoped
context, store concise decisions and evidence, and delegate only when the
benefit exceeds coordination cost.

### Durable, concise truth

Persist decisions, approved outcomes, task status, verification evidence, and
short summaries. Do not persist raw conversations, hidden reasoning, or
duplicate specifications as shared state.

### Runtime-neutral core

The workflow is portable across supported coding runtimes. Runtime-specific
adapters may explain installation or delegation capabilities, but they do not
change Cassor's authority, approval, or evidence rules.

## Current implementation

The following describes implemented behavior in the Go CLI and its automated
tests as of this document's review date.

### Local state and command surface

- `cassor init` discovers a Git repository and creates `.cassor/` state.
- SQLite stores configurable categories, Epics, roadmap items, plan revisions,
  tasks, lifecycle records, acceptance criteria, evidence, feature
  relationships, and portability data.
- The CLI supports Epic, item, plan, task, lifecycle, context, audit, validation,
  skill, export/import, benchmark, and static-site commands.
- State is repository-local. `cassor export` and `cassor import` provide a
  versioned, deterministic transfer and recovery boundary.

### Safe delivery workflow

The persisted workflow currently uses a roadmap item, approved plan revision,
and task hierarchy. Plans require meaningful acceptance criteria and task-level
verification requirements. Only one approved revision is active for an item;
completion checks active-plan tasks and required criteria. Verification and
waiver evidence are retained, and blocked tasks can resume with their history.

Features may have zero or one Epic parent. Epics are optional grouping records:
they do not own plans or tasks, and Cassor does not infer Epic completion from
child Feature state.

### Context management

`cassor context` provides a compact navigation summary. A scoped JSON context
can be requested for one item, role, and maximum byte size. It includes the
active assignment, applicable criteria, next action, and stable references for
omitted detail. This is the primary mechanism for resuming work without
reloading a full project history.

### Skills and runtime adapters

Cassor embeds one portable `cassor` skill plus adapters for Codex, Claude Code,
and GitHub Copilot. Installations are project-scoped, tracked by a manifest,
and protected from overwriting local modifications. Model-routing configuration
is optional runtime metadata.

Cassor does **not** create or control subagents itself. The host runtime owns
agent creation, isolation, model availability, and execution. Cassor provides
the workflow contract and context those agents use. When delegation is absent
or wasteful, the orchestrator follows the same contract locally.

### Projections and diagnostics

Cassor can generate a static documentation and roadmap site. The source
repository additionally contains an enhanced SvelteKit documentation view.
These are derived views, not authoritative state. `cassor check` validates
state and generation invariants. Benchmark and run-report commands report only
the measurements supplied or observable for a run. Optional feature reports
can persist concise, observed comparison records keyed by an explicit
comparison key; unavailable values remain unavailable, and Cassor does not
estimate tokens, store transcripts, or produce rankings.

## Intended product model

The user-facing model should be easy to explain:

```text
Epic (optional)
└── Feature
    └── Task
```

- An **Epic** groups related outcomes when a request is too large for one
  feature. It is optional; Cassor must remain comfortable for small work.
- A **Feature** records the outcome, concise context, decisions, acceptance
  criteria, and approved execution plan.
- A **Task** is a bounded unit of investigation, implementation, or
  verification with a clear completion signal.

The persisted model supports this hierarchy while retaining the current plan
and task guarantees for Features. Existing data remains valid as ungrouped
Features after migration.

## Interaction model

```text
Request
  → focused discovery
  → concise plan and approval
  → bounded implementation and verification
  → compact outcome and resumable state
```

For independent or specialized work, the orchestrator may delegate a bounded
assignment to a subagent. Each assignment receives only the task-relevant
context and returns a concise finding, implementation result, or verification
result. The orchestrator remains responsible for user communication, scope,
approval, and the final decision.

The seven SDD-lite phases—Intake, Explore, Define, Plan, Implement, Verify,
and Record—remain useful internal guidance. They must not become mandatory
documents or visible ceremony for every task. The normal human experience is:

```text
Understand → Approve → Execute → Verify
```

Plan packets may retain structurally valid `open_questions` during Understand,
so the orchestrator can consolidate material uncertainty before approval. Plan
recording clears those questions by default: `cassor plan record` rejects a
packet that still has them. An explicitly approved exception may use
`--allow-open-questions` together with both `--approve` and
`--approved-by-user`; the CLI emits a conspicuous warning and preserves the
questions in the immutable recorded plan content.

## Explicit non-goals

Cassor is not:

- a replacement for issue trackers, portfolio planning, or team project
  management tools;
- a hosted, synchronized, multi-repository or multi-writer service;
- an autonomous coding system that bypasses meaningful human approval;
- a store for chat transcripts or agent reasoning;
- a requirement that every task uses subagents or every change produces a long
  specification; or
- a promise of token savings without comparable, runtime-observable evidence.

## Source-of-truth rules

For the current product state, precedence is:

1. Executable code and automated tests define what Cassor does today.
2. This PRD defines the approved product identity, principles, and intended
   direction.
3. `.cassor/` holds repository-specific workflow state.
4. Generated sites, exports, and summaries are projections or transfer
   artifacts, never competing product truth.

If a future implementation conflicts with this PRD, update the PRD through an
approved product change in the same delivery work.

## Product success

Cassor is succeeding when a user can:

- understand and approve a feature plan in minutes rather than read a large
  specification;
- resume a task using a small scoped context rather than replay a conversation;
- trust that agents will not confuse a proposal with approved work;
- choose delegation only when it improves the work; and
- inspect a concise record of outcomes and verification after delivery.

The next product evolution should focus on making the Epic → Feature → Task
model natural in the skill and CLI while preserving the existing guarantees for
approval, verification, portability, and bounded context.

## Evidence in this repository

Current-state claims are grounded in:

- `internal/cmd/` for the executable command surface;
- `internal/store/` and `internal/migrations/` for workflow, context,
  verification, and portability invariants;
- `internal/skills/` and `skills/portable/cassor/` for portable skill and
  adapter behavior; and
- the corresponding `*_test.go` files for automated behavioral coverage.
