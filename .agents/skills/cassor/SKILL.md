---
name: cassor
description: Coordinate repository or greenfield software work through focused discovery, incremental user questions, explicit plan approval, bounded implementation, verification, and Cassor roadmap state. Use when the user asks to plan and build a feature, fix, refactor, migration, or new software project with Cassor. Do not use for explanation-only requests or when the user explicitly asks to bypass Cassor.
---

# Cassor

Locate the Cassor project before acting. Run `cassor context --json` to load compact durable state and `cassor check` before relying on that state. When `cassor` is unavailable, use the runtime adapter's documented command fallback.

Follow the SDD-lite lifecycle: Intake → Explore → Define → Plan → Implement → Verify → Record. Define acceptance criteria before planning; Verify records manual or automated evidence, and a feature cannot complete while any criterion is unresolved. The Cassor CLI—not prompt text—owns legal persisted-state transitions.

Inspect discoverable facts before asking questions. Select only the role lenses justified by the request; read the matching role contract before dispatching work. Ask meaningful questions incrementally, including recommendations and material consequences when appropriate. Before presenting a plan, confirm that relevant coverage is answered, safely inferred, or irrelevant.

Present a compact plan packet with no material open questions. Do not mutate repository files or persistent Cassor state until the user explicitly approves the exact plan revision, except for a user-requested, non-logic, mechanically safe correction with no behavior, API, schema, configuration, workflow, or UI impact. After approval, record the packet atomically with `cassor plan record --file … --approve --approved-by-user`, then execute approved tasks sequentially. Run relevant verification before marking each task complete.

Tactical implementation changes inside the approved behavior are allowed. Stop for an amendment before changing user-visible behavior, scope, architecture, public interfaces, persisted data, dependencies, compatibility, acceptance criteria, migrations, destructive behavior, or delivery outcomes. Repairs that remain inside the approved plan may continue.

When runtime capabilities are limited, perform the role contracts sequentially and compact findings immediately. Never claim a worker model or isolated worker context that the runtime cannot guarantee.

## Codex adaptive orchestration

Before dispatch, read `.cassor/agents.toml` and assess only capabilities the current Codex runtime actually exposes: subagents, parallel execution, isolated worker contexts, compact structured returns, and explicit worker-model routing. Do not infer or claim any unavailable capability. Explicit model routing is optional: when available, select the model mapped for the bounded role in `[runtimes.codex]`; when unavailable or the mapping is invalid, use the runtime default without claiming a model choice.

The default Codex mapping is `gpt-5.6-sol` for the `orchestrator` and `discovery` roles, and `gpt-5.6-luna` for `implementation` and `verification`. The orchestrator and research workers therefore use Sol; the one approved implementation worker and independent verifier use Luna. Treat project configuration as authoritative, so users may substitute runtime-supported identifiers. Do not spawn a worker if its required capability, isolation, or mutation authority is unavailable.

Choose the execution shape from the work and those capabilities:

| Work shape | Execution |
| --- | --- |
| Small, local, tightly coupled, or not materially improved by specialization | Orchestrator performs the role locally. |
| Independent, bounded discovery lenses with material benefit, and safe parallel isolated contexts | Parallel read-only discovery workers. |
| Approved implementation with material benefit from delegation and safe isolation | Exactly one implementation worker; never concurrent implementation workers. |
| Substantial, independently checkable verification with safe parallel isolated contexts | An independent read-only verification worker, optionally parallel only with other read-only work. |
| Any required worker, parallel, isolation, or return capability is unavailable or unsafe | Perform the same role contracts sequentially in the orchestrator context. |

Parallel workers are read-only. An implementation worker receives write authority only after plan approval, only for its exact task and repository boundary, and must run alone with respect to repository mutations. The orchestrator is the only user-facing role and the only role allowed to record plans, transition Cassor state, or approve work. Every worker receives an exact bounded task, the relevant role contract, repository boundaries, explicit mutation authority, and its selected model only when runtime routing is supported. Workers must not ask the user questions, approve plans, alter Cassor state, or persist transcripts. Return only the applicable compact schema; if structured returns are unavailable, consolidate an equivalently compact, evidence-based result without representing it as runtime-provided structure.

At the end of a run, include a compact handoff summary: roles dispatched, execution mode (single-agent, sequential, or parallel), elapsed time when available, and token usage only when the runtime exposes it. Use `cassor run-report` for a shareable text or JSON rendering of observed metrics; unavailable token fields remain unavailable. Do not persist this summary, telemetry, or raw worker transcripts.

Read [the protocol](references/protocol.md) for lifecycle and approval detail, [schemas](references/schemas.md) when producing or consuming structured results, and a role contract as needed: [discovery](references/roles/discovery.md), [implementation](references/roles/implementation.md), or [verification](references/roles/verification.md).
