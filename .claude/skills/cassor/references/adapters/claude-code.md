# Claude Code adapter

This file is runtime-specific guidance for a project installation under
`.claude/skills/cassor`. The portable skill and Cassor CLI remain authoritative.

## Capability boundary

Treat Claude Code worker and model-routing capabilities as unavailable unless
the current session explicitly exposes and safely scopes them. A project skill
installation does not prove that subagents, isolation, or model selection are
available.

When those capabilities are unavailable, continue as one orchestrator using
the proportionate SDD-lite lifecycle. Preserve explicit plan approval,
amendment handling, bounded writes, verification, and Cassor CLI state
transitions.

Optional Claude configuration or custom worker profiles are not installed by
Cassor's default project adapter. Add them only through a separately approved
runtime-specific change, and keep them outside the canonical portable skill.

## Safe installation behavior

- Project assets belong under `.claude/skills/cassor`.
- Do not modify global Claude configuration, hooks, MCP settings, or unrelated
  instructions.
- Let `cassor skills status`, `cassor skills update`, and `cassor skills doctor`
  report managed-file drift before changing an installation.
- If the destination exists without a Cassor manifest entry, stop and resolve
  the conflict instead of overwriting it.
