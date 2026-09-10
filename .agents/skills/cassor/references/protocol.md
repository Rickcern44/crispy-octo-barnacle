# Cassor Protocol

## Roles and authority

The orchestrator is the only role that interacts with the user. It selects
proportionate discovery lenses, consolidates compact findings, asks questions,
obtains approval, and records accepted decisions through Cassor commands.

Workers never ask the user questions, approve work, expand scope, persist
speculative roadmap items, or mutate Cassor state outside their assigned
authority. Do not persist worker transcripts or hidden reasoning.

## Adaptive dispatch

Load `adaptive-orchestration.md` only when delegation is useful. It contains
capability assessment, execution-shape selection, runtime/model routing,
compact worker returns, and non-persistent reporting guidance. Keep small,
local, sequentially dependent, or tightly coupled work in the orchestrator
context.

## Efficiency reporting

Use `cassor run-report` to render a compact, non-persistent record of metrics
that the runtime actually exposes. Record execution mode, dispatched roles,
elapsed time, observed tool calls, and verification result. Supply token flags
only when the runtime returns those metrics. Missing values are reported as
unavailable; never estimate them or treat them as zero.

## Understand

Inspect repository facts before questions. Consolidate only material questions,
recommendations, boundaries, and acceptance signals. Use only relevant
repository, requirements, technical, risk, product, or greenfield lenses.
Recommendations and future opportunities require explicit user approval before
entering durable roadmap state.

Cover the desired outcome, users and permissions, current behavior, included
and excluded scope, edge cases, compatibility, UX/API contract, failure
behavior, testing, migration, delivery, security, and data integrity when
relevant. Ask whether any intended behavior or constraint remains uncovered
before preparing the plan.

For a small local change, use the proportionate path: inspect context and
repository facts, state the bounded outcome and acceptance check, obtain the
same explicit plan approval, implement locally, verify the result, and record
the evidence. Do not create worker findings, broad discovery artifacts, or
benchmark output unless they add decision or verification value.

## Approve → Execute → Verify

The user approves an immutable plan revision, not individual commands or diffs.
One approval may authorize multiple tasks. Record the packet only after
approval with both `--approve` and `--approved-by-user`; unresolved
`open_questions` are rejected by default. The explicit
`--allow-open-questions` override is permitted only with both approval flags,
must emit a conspicuous warning, and preserves the questions in immutable plan
content. Include an approval note only when useful. Do not store chat
transcripts.

Start one approved task at a time. Implement within the plan boundary, run its
stated verification, and mark it complete with concise evidence. If blocked,
mark it blocked with the reason. Build the roadmap after Cassor state changes
and run `cassor check` before reporting final completion.

## Amendments and repair

Request an amendment for material divergence. Create and obtain approval for a
new plan revision before continuing divergent work. A failed verification may
be repaired without renewed approval only when the repair preserves the
approved behavior, scope, and constraints.
