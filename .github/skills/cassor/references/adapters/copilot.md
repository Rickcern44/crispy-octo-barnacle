# GitHub Copilot adapter

This file is runtime-specific guidance for a project installation under
`.github/skills/cassor`. The portable skill and Cassor CLI remain authoritative.

## Capability boundary

Treat Copilot worker delegation and explicit model routing as unavailable unless
the current session explicitly exposes those capabilities. The presence of an
agent skill directory does not establish that a worker can safely mutate the
repository or that a selected model is active.

When delegation is unavailable, continue as one orchestrator using the
proportionate SDD-lite lifecycle. Preserve explicit plan approval, amendment
handling, bounded writes, verification, and Cassor CLI state transitions.

Copilot instructions, custom agents, hooks, and repository-wide settings are
not installed by Cassor's default project adapter. Add them only through a
separately approved runtime-specific change, and keep them outside the
canonical portable skill.

## Safe installation behavior

- Project assets belong under `.github/skills/cassor`.
- Do not modify repository-wide Copilot instructions, custom agents, hooks,
  MCP settings, or unrelated files.
- Let `cassor skills status`, `cassor skills update`, and `cassor skills doctor`
  report managed-file drift before changing an installation.
- If the destination exists without a Cassor manifest entry, stop and resolve
  the conflict instead of overwriting it.
