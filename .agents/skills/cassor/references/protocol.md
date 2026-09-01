# Cassor Protocol

## Roles and authority

The orchestrator is the only role that interacts with the user. It selects proportionate discovery lenses, consolidates compact findings, asks questions, obtains approval, and records accepted decisions through Cassor commands.

Workers never ask the user questions, approve work, expand scope, persist speculative roadmap items, or mutate Cassor state outside their assigned authority. Do not persist worker transcripts or chain-of-thought.

## Adaptive Codex dispatch

Before dispatch, the orchestrator records a transient capability assessment for the current Codex runtime: subagents, parallel execution, isolated worker contexts, compact structured returns, and explicit worker-model routing. Treat an unknown capability as unavailable. Do not claim a worker model, isolated context, parallelism, structured return, metric, or telemetry that the runtime has not exposed.

Dispatch is justified only when the work is both independent and bounded, and specialization or concurrency materially improves the result. Keep small, local, sequentially dependent, or tightly coupled work in the orchestrator context. Use this decision policy:

1. If safe subagents are unavailable, perform each relevant role contract sequentially in the orchestrator context.
2. Parallel work requires available parallel execution and isolated contexts. In a shared workspace, every parallel worker is read-only. If either prerequisite is unavailable, do not parallelize; use the same role contracts sequentially.
3. Use focused discovery workers only for separate read-only lenses with a material benefit. Do not split one tightly coupled investigation merely to create workers.
4. Delegate implementation only after the user has approved and Cassor has recorded the exact plan revision. Grant one implementation worker explicit, task-scoped write authority. Never run more than one implementation worker, and never overlap its repository mutations with another worker.
5. Use an independent verification worker only for substantial, independently checkable verification. It is read-only and may run in parallel only with other read-only work that meets the parallel prerequisites.
6. Structured returns are preferred for dispatched work. If the runtime cannot provide them, either collect a compact evidence-based result in the applicable schema format or execute the role sequentially; never label an ordinary response as a runtime-structured return.
7. Explicit worker-model routing is optional. If the runtime supports it, choose a model appropriate to the bounded role and report only that routing was available and used. If it does not, use the runtime default and do not imply a specific model was selected.

A worker receives only its exact task, relevant state, role contract, repository boundary, and explicit mutation authority. Implementation writes require approved, recorded plan authority; discovery and verification remain read-only. Workers never ask the user questions, approve work, record plans, transition Cassor state, or persist transcripts. The orchestrator alone retains those authorities.

Aggregate only compact, evidence-based findings in the applicable schema. The final handoff reports dispatched roles and execution mode; include elapsed time or token usage only when the runtime exposes them. None of this telemetry is persisted.

## Efficiency reporting

Use `cassor run-report` to render a compact, non-persistent record of metrics that the runtime actually exposes. Record execution mode, dispatched roles, elapsed time, observed tool calls, and verification result. Supply token flags only when the runtime returns those metrics. Missing values are reported as unavailable; never estimate them or treat an unavailable value as zero.

To evaluate adaptive orchestration, run a representative task twice from comparable repository state: once with a single agent and once with the chosen worker configuration. Compare acceptance-criterion pass rate first, then elapsed time, tool calls, review churn, and runtime-reported token usage when available. Do not persist raw transcripts, hidden reasoning, or run telemetry in Cassor state.

## Discovery and planning

Inspect repository facts before questions. Use only relevant lenses: repository, requirements, technical, risk, product, or greenfield domain. Classify findings as required, recommended, or future opportunity. Recommendations and future opportunities require explicit user approval before entering durable roadmap state.

Cover the desired outcome, users and permissions, current behavior, included and excluded scope, edge cases, compatibility, UX/API contract, failure behavior, testing, migration, delivery, security, and data integrity when relevant. Ask whether any intended behavior or constraint remains uncovered before preparing the plan.

## Approval and execution

The user approves an immutable plan revision, not individual commands or diffs. One approval can authorize multiple tasks. Record the packet only after approval with both `--approve` and `--approved-by-user`; include an approval note only when useful. Do not store chat transcripts.

Start one approved task at a time. Implement within the plan boundary, run its stated verification, and mark it complete with concise evidence. If blocked, mark it blocked with the reason. Build the roadmap after Cassor state changes and run `cassor check` before reporting final completion.

## Amendments and repair

Request an amendment for material divergence. Create and obtain approval for a new plan revision before continuing divergent work. A failed verification may be repaired without renewed approval only when the repair preserves the approved behavior, scope, and constraints.
