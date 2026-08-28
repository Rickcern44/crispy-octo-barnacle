---
name: cassor
description: Coordinate repository or greenfield software work through focused discovery, incremental user questions, explicit plan approval, bounded implementation, verification, and Cassor roadmap state. Use when the user asks to plan and build a feature, fix, refactor, migration, or new software project with Cassor. Do not use for explanation-only requests or when the user explicitly asks to bypass Cassor.
---

# Cassor

Locate the Cassor project before acting. Run `cassor context --json` to load compact durable state and `cassor check` before relying on that state. When `cassor` is unavailable, use the runtime adapter's documented command fallback.

Follow the conceptual state flow: intake → inspect → clarify → plan-draft → awaiting-approval → implementing → verifying → completed. The Cassor CLI—not prompt text—owns legal persisted-state transitions.

Inspect discoverable facts before asking questions. Select only the role lenses justified by the request; read the matching role contract before dispatching work. Ask meaningful questions incrementally, including recommendations and material consequences when appropriate. Before presenting a plan, confirm that relevant coverage is answered, safely inferred, or irrelevant.

Present a compact plan packet with no material open questions. Do not mutate repository files or persistent Cassor state until the user explicitly approves the exact plan revision. After approval, record the packet atomically with `cassor plan record --file … --approve --approved-by-user`, then execute approved tasks sequentially. Run relevant verification before marking each task complete.

Tactical implementation changes inside the approved behavior are allowed. Stop for an amendment before changing user-visible behavior, scope, architecture, public interfaces, persisted data, dependencies, compatibility, acceptance criteria, migrations, destructive behavior, or delivery outcomes. Repairs that remain inside the approved plan may continue.

When runtime capabilities are limited, perform the role contracts sequentially and compact findings immediately. Never claim a worker model or isolated worker context that the runtime cannot guarantee.

Read [the protocol](references/protocol.md) for lifecycle and approval detail, [schemas](references/schemas.md) when producing or consuming structured results, and a role contract as needed: [discovery](references/roles/discovery.md), [implementation](references/roles/implementation.md), or [verification](references/roles/verification.md).
