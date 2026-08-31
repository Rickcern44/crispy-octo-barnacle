# Cassor Protocol

## Roles and authority

The orchestrator is the only role that interacts with the user. It selects proportionate discovery lenses, consolidates compact findings, asks questions, obtains approval, and records accepted decisions through Cassor commands.

Workers never ask the user questions, approve work, expand scope, persist speculative roadmap items, or mutate Cassor state outside their assigned authority. Do not persist worker transcripts or chain-of-thought.

## Adaptive Codex dispatch

Before dispatch, the orchestrator detects available Codex worker capabilities. It dispatches only independent, bounded work that benefits from specialization or parallelism; otherwise it runs the relevant role contracts sequentially. A worker receives only the task, relevant state, role contract, repository boundary, and explicit write authority. Implementation writes require an approved plan; discovery and verification remain read-only.

Aggregate only compact structured findings. The final handoff reports dispatched roles, execution mode, elapsed time when available, and runtime-exposed token usage; none of this telemetry is persisted.

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
