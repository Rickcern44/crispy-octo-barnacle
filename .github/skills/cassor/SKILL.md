---
name: cassor
description: Coordinate repository or greenfield software work through focused discovery, incremental user questions, explicit plan approval, bounded implementation, verification, and Cassor roadmap state. Use when the user asks to plan and build a feature, fix, refactor, migration, or new software project with Cassor. Do not use for explanation-only requests or when the user explicitly asks to bypass Cassor.
---

# Cassor

Cassor is the workflow authority for repository work. First locate the project,
run `cassor context --json`, and run `cassor check` before relying on durable
state. If Cassor is unavailable, use the runtime adapter's documented fallback.

The visible lifecycle is:

`Understand → Approve → Execute → Verify`

Keep the work proportionate: small, well-bounded work needs only focused
inspection and a compact plan; expand discovery only for material uncertainty,
risk, or compatibility impact.

Understand by inspecting discoverable facts with only the relevant repository,
requirements, technical, risk, product, or greenfield lens. Consolidate only
material questions, recommendations, boundaries, and acceptance signals. Read
the matching role contract before dispatching work. Questions that remain are
structurally valid in a plan packet, but normal recording requires them to be
cleared.

The orchestrator is the only user-facing authority. Workers do not ask the
user questions, approve plans, expand scope, alter Cassor state, or persist raw
transcripts. The CLI owns legal persisted transitions: do not substitute prompt
text for `cassor` validation.

Do not mutate repository files or Cassor state until the user approves the
exact plan revision. After approval, record it atomically with
`cassor plan record --file … --approve --approved-by-user`; unresolved
`open_questions` are rejected by default. The explicit
`--allow-open-questions` override is legal only alongside both approval flags,
emits a conspicuous warning, and preserves the questions in immutable plan
content. Execute approved tasks sequentially, verify each task, and record
concise evidence. Request an amendment before changing behavior, scope,
architecture, interfaces, schemas, dependencies, compatibility, acceptance
criteria, migrations, destructive behavior, or delivery outcomes. Repairs
inside the approved envelope may continue.

When delegation would materially help, load
`references/adaptive-orchestration.md`; otherwise perform the role contracts
locally and sequentially. Load `references/protocol.md`, `references/schemas.md`,
and the role contract only when their detail is needed. Use
`cassor run-report` for observed, non-persistent run metrics; unavailable
telemetry remains unavailable.
