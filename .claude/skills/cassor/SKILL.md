---
name: cassor
description: Coordinate repository or greenfield software work through focused discovery, incremental user questions, explicit plan approval, bounded implementation, verification, and Cassor roadmap state. Use when the user asks to plan and build a feature, fix, refactor, migration, or new software project with Cassor. Do not use for explanation-only requests or when the user explicitly asks to bypass Cassor.
---

# Cassor

Cassor is the workflow authority for repository work. First locate the project,
run `cassor context --json`, and run `cassor check` before relying on durable
state. If Cassor is unavailable, use the runtime adapter's documented fallback.

Follow the proportionate SDD-lite lifecycle:

`Intake → Explore → Define → Plan → Implement → Verify → Record`

Use the phases as guidance, not mandatory paperwork. Small local changes may
use compact discovery and a short plan, but every change still needs execution
readiness, explicit approval before mutation, amendment handling for material
divergence, and verification before completion.

Inspect discoverable facts before asking questions. Select only the relevant
repository, requirements, technical, risk, product, or greenfield lens. Read
the matching role contract before dispatching work. Ask only decisions that
inspection cannot resolve, then produce a compact plan with acceptance criteria
and no unresolved material questions.

The orchestrator is the only user-facing authority. Workers do not ask the
user questions, approve plans, expand scope, alter Cassor state, or persist raw
transcripts. The CLI owns legal persisted transitions: do not substitute prompt
text for `cassor` validation.

Do not mutate repository files or Cassor state until the user approves the
exact plan revision. After approval, record it atomically with
`cassor plan record --file … --approve --approved-by-user`, execute approved
tasks sequentially, verify each task, and record concise evidence. Request an
amendment before changing behavior, scope, architecture, interfaces, schemas,
dependencies, compatibility, acceptance criteria, migrations, destructive
behavior, or delivery outcomes. Repairs inside the approved envelope may
continue.

When delegation would materially help, load
`references/adaptive-orchestration.md`; otherwise perform the role contracts
locally and sequentially. Load `references/protocol.md`, `references/schemas.md`,
and the role contract only when their detail is needed. Use
`cassor run-report` for observed, non-persistent run metrics; unavailable
telemetry remains unavailable.
