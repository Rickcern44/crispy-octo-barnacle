# Adaptive orchestration reference

Load this reference only when delegation is useful for the current task. The
entry skill and protocol remain authoritative for approval, scope, and state.

## Capability assessment

Before dispatch, inspect `.cassor/agents.toml` and assess only capabilities the
current runtime actually exposes:

- subagent creation;
- parallel execution;
- isolated worker contexts;
- compact structured returns; and
- explicit worker-model routing.

Treat unknown capabilities as unavailable. Never claim a worker model,
isolation boundary, parallel execution, structured return, metric, or telemetry
that the runtime has not exposed.

## Execution shape

| Work shape | Execution |
| --- | --- |
| Small, local, tightly coupled, or not materially improved by specialization | Orchestrator performs the role locally. |
| Independent, bounded discovery lenses with safe isolated parallel contexts | Parallel read-only discovery workers. |
| Approved implementation with material delegation benefit and safe isolation | Exactly one implementation worker. |
| Substantial, independently checkable verification with safe isolation | Independent read-only verification worker, optionally alongside other read-only work. |
| Any required capability, isolation, or return guarantee is unavailable | Perform the same role contracts sequentially in the orchestrator context. |

Parallel workers are read-only. An implementation worker receives write
authority only after the exact plan revision is approved and recorded, only for
its assigned task and repository boundary, and never concurrently with another
repository mutator. Verification workers remain read-only.

## Runtime and model routing

Portable role classes use capability labels rather than vendor model names:

```toml
[models]
orchestrator = "frontier"
discovery = "frontier"
implementation = "economical-coding"
verification = "economical-coding"
```

If the runtime supports explicit routing and `.cassor/agents.toml` provides a
valid `[runtimes.codex]` mapping, use the configured identifiers. If routing is
unavailable or incomplete, use the runtime default without implying that a
specific model was selected. The current project defaults are configuration,
not a guarantee of runtime availability.

## Compact returns and reporting

Workers return only the applicable compact result schema from
`references/schemas.md`, with paths, commands, and evidence. If structured
returns are unavailable, consolidate an equivalently compact evidence-based
result without representing it as runtime-provided structure.

Use `cassor run-report` to render execution mode, roles, elapsed time, observed
tool calls, verification, context packet counts/bytes, handoffs, and
runtime-reported token values. `cassor run-report record` may persist these
concise observed metrics against a feature, and `cassor run-report compare
--comparison-key <key>` shows only records sharing that explicit key. Missing
fields are explicitly unavailable, never estimated or treated as zero. Do not
persist raw worker transcripts, hidden reasoning, or inferred rankings.
