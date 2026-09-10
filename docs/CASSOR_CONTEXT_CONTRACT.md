# Cassor context contract

This operational contract supports the [Cassor PRD](PRD.md)'s token-efficiency
principle.

`cassor context` returns a bounded navigation summary with counts and stable
commands for the first ten entries in each queue. It does not include plan
content, task descriptions, or the backlog payload.

Use scoped context to resume an item:

```text
cassor context --item 17 --role implementation --max-bytes 12000 --json
```

Supported roles are `implementation`, `verification`, and `planning`. The
response identifies the item, active approved plan revision, packet goal and
constraints, revision-scoped criteria, the next task and its verification
requirements, blockers, and evidence references. Every omitted detail includes
an ID and a command for retrieving it with `cassor plan show`, `cassor task
show`, or the lifecycle criterion command.

`max-bytes` measures compact JSON bytes. Required fields remain present. If
optional detail cannot fit, `truncated` is true and `omitted_details` lists
stable references. If the bound is smaller than the required projection, the
command fails with the minimum size instead of returning incomplete execution
constraints. Mutation commands return IDs, resulting state, and a next action
in their default text output; pass `--json` where the full recorded artifact is
needed explicitly.
