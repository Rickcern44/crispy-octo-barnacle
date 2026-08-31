# Cassor Skills Specification

Status: Draft for implementation planning  
Project: `github.com/rickcern44/cassor`  
CLI: `cassor`

## 1. Purpose

Cassor provides a lightweight, agent-agnostic software-development workflow with deliberate discovery, explicit human approval, bounded implementation, independent verification, and durable low-context project state.

This document specifies the skills layer that will sit on top of the Cassor CLI. The CLI and SQLite database remain authoritative. Skills teach an agent when and how to use the CLI and how to coordinate specialized workers without allowing planning transcripts or speculative work to expand shared context.

The skills layer must support:

- OpenAI Codex
- Anthropic Claude Code
- GitHub Copilot

The design uses the open `SKILL.md` structure shared by these runtimes. Runtime-specific adapters handle installation paths, subagent configuration, model selection, and capability differences.

## 2. Non-goals

The skills layer does not:

- Replace Cassor's SQLite state with Markdown files.
- Allow an agent to approve its own roadmap item or plan.
- Guarantee identical subagent behavior across runtimes.
- Require every task to spawn multiple agents.
- Persist full worker transcripts.
- Turn small changes into multi-document specifications.
- Install hooks or edit unrelated agent configuration without explicit authorization.
- Provide a general-purpose issue tracker or multi-repository portfolio manager.

## 3. Design principles

### 3.1 One portable workflow

Cassor must have one canonical, runtime-neutral skill source. Codex, Claude Code, and Copilot adapters are generated or copied from that source. The workflow must not be independently rewritten three times.

### 3.2 One user-facing skill

The default installation exposes one discoverable skill named `cassor`. Discovery, implementation, and verification are internal role contracts rather than separate user-facing skills.

This prevents accidental partial invocation such as implementing a plan without loading the approval and scope rules.

### 3.3 The orchestrator owns the conversation

Only the orchestrator interacts with the user. Workers inspect, analyze, implement, or verify and return compact structured results to the orchestrator.

Workers must not:

- Ask the user questions directly.
- Approve work.
- Expand scope.
- Persist speculative roadmap entries.
- Modify Cassor state outside their assigned permissions.

### 3.4 Proportional process

The orchestrator selects only the planning lenses justified by the task. A small localized bug may use one repository investigator. A greenfield feature with security and data implications may use several planners.

There is no fixed number of planners, questions, documents, or tasks.

### 3.5 Compact durable state

SQLite is authoritative. Agents retrieve a bounded view through `cassor context --json` and mutate state through Cassor commands.

Persistent state should contain decisions, approved outcomes, task status, verification evidence, and accepted roadmap entries. It should not contain full chain-of-thought, chat transcripts, or duplicate specifications.

### 3.6 Efficiency measurement

Cassor can render an ephemeral run report with `cassor run-report`. The report records caller-supplied execution mode, roles, elapsed time, tool-call count, verification result, and token metrics only when the runtime exposes them. Missing token values are explicitly unavailable, never estimated or treated as zero. Reports are not persisted in SQLite.

Evaluate orchestration with repeatable A/B runs from comparable repository state: a single-agent baseline and an adaptive worker configuration. Compare acceptance-criterion pass rate first, then elapsed time, tool calls, review churn, and API-reported input, cached input, output, reasoning, and total tokens where available. Runtime-hosted sessions without usage telemetry can still be compared on quality and elapsed work, but must not claim a token saving.

## 4. Canonical package layout

The Cassor source repository should contain:

```text
skills/
├── portable/
│   └── cassor/
│       ├── SKILL.md
│       └── references/
│           ├── protocol.md
│           ├── schemas.md
│           └── roles/
│               ├── discovery.md
│               ├── implementation.md
│               └── verification.md
└── adapters/
    ├── codex/
    ├── claude-code/
    └── copilot/
```

Only create adapter files that a runtime actually requires. Do not add placeholder files or duplicate the portable references.

The portable skill and required adapter assets should be embedded into the Go binary with `go:embed` for installation into other repositories.

## 5. Portable `SKILL.md`

The skill frontmatter should be concise and discriminating:

```yaml
---
name: cassor
description: Coordinate repository or greenfield software work through focused discovery, incremental user questions, explicit plan approval, bounded implementation, verification, and Cassor roadmap state. Use when the user asks to plan and build a feature, fix, refactor, migration, or new software project with Cassor. Do not use for explanation-only requests or when the user explicitly asks to bypass Cassor.
---
```

The body should contain only:

1. How to locate and validate the Cassor project.
2. How to load compact context.
3. The workflow state machine.
4. The approval invariants.
5. How to select relevant role references.
6. How to handle runtime capability limitations.
7. Links to the protocol, schemas, and role contracts.

Detailed output schemas and role instructions belong in references so they are loaded only when needed.

## 6. Workflow state machine

The portable protocol uses these conceptual states:

```text
intake
  → inspect
  → clarify
  → plan-draft
  → awaiting-approval
  → implementing
  → verifying
  → completed
```

Additional transitions:

```text
implementing → amendment-required → clarify|plan-draft
verifying    → repair             → implementing
verifying    → amendment-required → clarify|plan-draft
any active state → blocked
```

The Cassor CLI owns legal persisted-state transitions. The skill must never treat prompt instructions as a substitute for CLI validation.

## 7. Approval contract

### 7.1 Plan approval

- No repository mutation begins until the user approves a concrete plan revision.
- Approval is attached to an immutable revision.
- One approval may authorize multiple tasks.
- Approved tasks execute sequentially and automatically.
- Each task is verified before it is marked complete.
- Concise progress updates do not require acknowledgment.

### 7.2 Tactical adaptation

An implementer may make tactical choices inside the approved envelope, including:

- Following existing repository naming and organization.
- Editing an additional file clearly required by the approved behavior.
- Selecting an equivalent private helper or implementation technique.
- Correcting an inaccurate planned file list.
- Adding a narrowly relevant test.

The implementer must return an amendment request before changing:

- User-visible behavior
- Scope
- Architecture or component boundaries
- Public interfaces or persisted data
- Dependencies
- Compatibility guarantees
- Acceptance criteria
- Migration or destructive behavior
- Approved delivery outcomes

### 7.3 Repair after verification

A repair that merely corrects an implementation mistake inside the approved plan may continue. A repair requiring a new approach or changed behavior requires a revised plan and new approval.

## 8. Orchestrator behavior

The orchestrator should use the most capable configured model class. It is responsible for:

- Understanding the user's outcome.
- Loading `cassor context --json`.
- Inspecting the repository before asking discoverable questions.
- Selecting planning lenses.
- Dispatching focused worker tasks when supported.
- Consolidating and deduplicating worker findings.
- Asking meaningful questions incrementally.
- Separating required work from recommendations and future opportunities.
- Producing a compact plan revision.
- Obtaining explicit user approval.
- Dispatching implementation and verification.
- Handling amendments and bounded repair.
- Updating durable Cassor state only after relevant user decisions.

The orchestrator must not expose raw worker transcripts to the user unless explicitly requested.

## 9. Discovery role

The orchestrator selects one or more lenses rather than spawning generic planners.

Supported lenses include:

- Repository: architecture, conventions, dependencies, affected files, existing tests.
- Requirements: observable behavior, users, permissions, edge cases, acceptance criteria.
- Technical: implementation approaches, tradeoffs, compatibility, migration, delivery.
- Risk: security, data integrity, destructive behavior, operational impact, rollback.
- Product: user flow, value, scope boundaries, deferred opportunities.
- Greenfield domain: smallest useful slice, domain concepts, interfaces, foundational choices.

Each discovery worker receives:

- The user request.
- Its assigned lens.
- The minimum relevant Cassor context.
- Repository boundaries and inspection permissions.
- A prohibition on repository mutation.
- The discovery-result schema.

Each discovery worker returns:

```yaml
summary: Concise lens-specific conclusion
findings:
  - evidence: Repository fact or explicit user statement
    implication: Why it matters
questions:
  - question: Decision the user needs to make
    reason: Why repository inspection cannot resolve it
    recommendation: Preferred answer, when justified
    consequences: Material tradeoffs
assumptions:
  - statement: Safely inferred fact
    confidence: high|medium|low
risks:
  - risk: Material risk
    mitigation: Suggested treatment
roadmap_proposals:
  - title: Proposed future item
    category: Suggested category
    rationale: Why it may matter
readiness:
  status: ready|questions-remain|blocked
  reason: Brief explanation
```

Workers must cite repository paths, commands, or user statements for material findings. They should not generate questions merely to increase the question count.

## 10. Question coverage gate

Before presenting a plan, the orchestrator determines whether each relevant area is answered, safely inferred, irrelevant, or unresolved:

- Desired outcome
- Users and permissions
- Current behavior
- Included and excluded scope
- Edge cases
- Compatibility
- UX or API contract
- Failure behavior
- Testing
- Migration
- Delivery and rollout
- Security and data integrity

Questions are asked incrementally as they become relevant. Each question should normally include the orchestrator's recommendation and the consequence of choosing differently.

Immediately before preparing the final plan, the orchestrator asks whether any intended behavior or constraint remains uncovered.

## 11. Roadmap proposal handling

Planner discoveries are classified as:

- Required: necessary for the requested outcome to work safely or correctly.
- Recommended: valuable enough to ask whether it should enter the current plan.
- Future opportunity: potentially useful but outside the current implementation.

These are classifications of findings, not fixed roadmap categories.

Persistent roadmap categories are project-configurable. Initial examples include Needed, Recommended, Stretch goal, Technical debt, Security, and Research.

No proposal is added to the active plan or persisted roadmap without explicit user approval. The orchestrator should batch low-urgency future opportunities near plan review rather than interrupting discovery for each one.

## 12. Plan packet

The orchestrator creates a compact plan revision containing:

```yaml
roadmap_item: RM-###
revision: 1
goal: Observable completed outcome
scope:
  included: []
  excluded: []
decisions: []
acceptance_criteria: []
constraints: []
tasks:
  - id: TASK-###
    title: Small verifiable unit
    depends_on: []
    expected_changes: []
    verification: []
risks: []
approved_roadmap_additions: []
open_questions: []
```

A plan cannot enter `awaiting-approval` while material open questions remain. The user approves the entire revision, not each task individually.

## 13. Implementation role

The implementation worker receives:

- The exact approved plan revision.
- One active task at a time.
- Relevant repository context.
- Explicit included and excluded scope.
- The adaptation boundary.
- Required verification commands.

It returns:

```yaml
task: TASK-###
status: completed|blocked|amendment-required
changes:
  - path: path/to/file
    summary: What changed
tactical_variances:
  - variance: Difference from the predicted mechanics
    reason: Why it remained inside the approved envelope
verification:
  - command: command run
    result: passed|failed|not-run
    evidence: Concise output or reason
amendment: null
```

When an amendment is required, the worker stops before implementing the divergent behavior and returns:

```yaml
discovery: What was found
impact: Why the approved plan cannot be followed
recommended_change: Proposed amendment
alternatives:
  - option: Alternative
    tradeoff: Consequence
affected_scope: []
approval_required: true
```

## 14. Verification role

Verification should be independent for substantial changes. Small low-risk changes may be verified by the orchestrator when runtime capacity or proportionality makes a separate worker unnecessary.

The verifier receives the approved revision, task results, repository diff, and acceptance criteria. It checks:

- Required behavior
- Test and build results
- Unintended scope expansion
- Compatibility promises
- Missing edge cases
- Data or migration safety
- Documentation only when required by the approved plan

It returns:

```yaml
status: passed|repairable|amendment-required|blocked
criteria:
  - criterion: Approved acceptance criterion
    result: passed|failed|not-verifiable
    evidence: Concise evidence
issues:
  - severity: blocking|important|advisory
    finding: What is wrong
    treatment: repair|amendment|roadmap-proposal|none
recommended_next_action: Complete, repair, amend, or block
```

Advisory improvements do not fail an otherwise conforming implementation and are not persisted without user approval.

## 15. Model classes and capability negotiation

The portable skill uses capability labels rather than vendor model names:

```yaml
models:
  orchestrator: frontier
  discovery: economical
  implementation: economical-coding
  verification: economical-coding
```

Each adapter maps these labels to runtime-supported configuration where possible.

At startup, the orchestrator determines whether the runtime supports:

- Subagent creation
- Explicit worker model selection
- Parallel workers
- Isolated worker context
- Custom worker profiles
- Worker-to-orchestrator structured return

Fallback rules:

1. If subagents and model selection are supported, use configured role models.
2. If subagents exist without model selection, use them with the runtime default.
3. If no subagents exist, execute role contracts sequentially in the orchestrator context and immediately compact their results.
4. Never claim that a cheaper model was used when the runtime cannot guarantee it.

## 16. Runtime adapters

### 16.1 Codex

Project skill location:

```text
.agents/skills/cassor/
```

User skill location:

```text
~/.agents/skills/cassor/
```

The Codex adapter should rely on the portable skill and add only Codex-specific delegation/model-routing instructions supported by the installed Codex version. It must not duplicate the protocol.

### 16.2 Claude Code

Project skill location:

```text
.claude/skills/cassor/
```

User skill location:

```text
~/.claude/skills/cassor/
```

The Claude adapter may install custom subagent definitions when explicitly requested and supported. Custom agents should preload only the role material they need. The installer must preserve existing Claude configuration.

### 16.3 GitHub Copilot

Preferred project skill location:

```text
.github/skills/cassor/
```

Supported personal locations include:

```text
~/.copilot/skills/cassor/
~/.agents/skills/cassor/
```

Copilot supports the open skill structure, but the adapter must treat subagent/model routing as optional runtime capability. It may add repository-wide Copilot instructions only with separate user authorization.

## 17. Installer commands

Required command surface:

```bash
cassor skills install --agent codex --scope project
cassor skills install --agent claude-code --scope project
cassor skills install --agent copilot --scope project
cassor skills install --agent all --scope project
cassor skills list
cassor skills status
cassor skills update
cassor skills uninstall
cassor skills doctor
```

Optional flags:

```text
--scope project|user
--dry-run
--json
--force
--with-workers
```

`--force` must never silently destroy customized files. It may replace only Cassor-managed files after displaying or returning the detected conflict and receiving explicit confirmation in interactive mode. Noninteractive forced replacement should require an additional explicit confirmation flag designed during implementation.

## 18. Installation ownership and manifest

Cassor should record installed artifacts in:

```text
.cassor/skills-installations.json
```

The manifest contains:

```yaml
schema_version: 1
installations:
  - agent: codex
    scope: project
    destination: .agents/skills/cassor
    cassor_version: 0.1.0
    installed_at: RFC3339 timestamp
    files:
      - path: SKILL.md
        sha256: content hash
```

The manifest tracks only files installed by Cassor. It must not claim ownership of pre-existing files.

Before install, update, or uninstall:

1. Resolve the repository root and requested scope.
2. Detect destination capabilities and conflicts.
3. Compare existing content with recorded hashes.
4. Produce a dry-run change set.
5. Refuse to overwrite locally modified managed files without explicit confirmation.
6. Write through a temporary directory and replace atomically where practical.
7. Update the manifest only after every requested operation succeeds.

Uninstall removes only unmodified Cassor-managed files. Modified files are preserved and reported.

## 19. `skills doctor`

`cassor skills doctor` validates:

- Cassor project discovery
- CLI/database compatibility
- Installed skill paths
- Required `SKILL.md` frontmatter
- Reference link integrity
- Manifest hashes
- Adapter/runtime compatibility
- Availability of configured worker mechanisms
- Whether installed artifacts are stale relative to the Cassor binary

It must distinguish errors from capability limitations. For example, lack of explicit worker model selection is a warning with a documented fallback, not necessarily an installation failure.

## 20. Security and authorization

- Skill installation is a repository or user-environment mutation and requires an explicit install command.
- The installer must not modify unrelated `AGENTS.md`, `CLAUDE.md`, Copilot instructions, hooks, MCP configuration, or global settings by default.
- Optional integration changes require separate flags and must appear in dry-run output.
- Workers inherit only the permissions required by their role.
- Discovery and verification are read-only unless a narrowly scoped verification command necessarily creates standard build artifacts.
- The implementation worker receives mutation authority only after plan approval.
- A skill prompt is guidance, not a security boundary; deterministic workflow rules must also be enforced by Cassor CLI transitions.

## 21. Testing strategy

### 21.1 Static validation

- Validate `SKILL.md` frontmatter.
- Verify every linked reference exists.
- Confirm adapter output contains the canonical portable content.
- Test deterministic rendering for each target path.

### 21.2 Installer tests

Use temporary repositories and home directories to test:

- Fresh install
- Repeated idempotent install
- Dry run
- Multiple agents
- Project and user scope
- Pre-existing conflicting files
- Locally modified managed files
- Safe update
- Safe uninstall
- Partial-failure rollback
- Windows-safe paths and file replacement behavior

### 21.3 Behavioral forward tests

Run realistic scenarios against each supported runtime when available:

1. Small existing-repository bug.
2. Ambiguous feature requiring multiple question rounds.
3. Greenfield application slice.
4. Implementation discovery requiring a material amendment.
5. Verification finding an in-scope repair.
6. Planner proposing useful out-of-scope work.
7. Runtime without subagent model selection.

Evaluate observable behavior rather than exact wording:

- No mutation before approval.
- Questions are meaningful and incremental.
- The plan remains compact.
- Worker results are synthesized rather than dumped.
- Material divergence causes an amendment.
- Roadmap proposals require approval.
- Each task is verified.

## 22. Implementation sequence

After the Cassor CLI contract is stable:

1. Create the canonical portable skill and references.
2. Add installer path resolution and manifests.
3. Implement Codex project-scope installation.
4. Validate Cassor against a realistic Codex task.
5. Implement Claude Code installation and optional worker profiles.
6. Validate Claude capability fallbacks.
7. Implement Copilot installation.
8. Validate project and user scopes.
9. Add update, uninstall, and doctor behavior.
10. Run cross-runtime behavioral scenarios and refine only from observed failures.

## 23. Initial product decisions captured

- The user interacts only with the orchestrator.
- The orchestrator uses the frontier model class.
- Workers prefer cheaper/faster model classes.
- Questions are incremental and coverage-driven, with no fixed maximum.
- Repository facts should be inspected rather than asked.
- The user approves every plan revision before repository mutation.
- An approved multi-task plan executes automatically task-by-task.
- Material changes require a plan amendment and renewed approval.
- Planners may recommend adjacent work, but persistence requires approval.
- Roadmap categories are configurable.
- SQLite is authoritative.
- Static HTML is a generated projection.
- One roadmap exists per repository.
- The initial supported runtimes are Codex, Claude Code, and GitHub Copilot.

## 24. Remaining decisions for implementation planning

These should be resolved when the skills milestone begins:

1. Exact Cassor CLI commands for creating and persisting plan revisions from structured worker output.
2. Whether project-scope is the default for `cassor skills install`. Recommended: yes.
3. Whether `--agent all` installs three copies or uses `.agents/skills` where multiple runtimes support it. Recommended initially: explicit per-runtime destinations for predictable discovery.
4. Whether optional custom worker profiles are installed by default. Recommended: no; require `--with-workers` until behavior is proven.
5. How runtime/model mappings are configured. Recommended: a project-level `.cassor/agents.toml` containing capability classes and optional runtime-specific overrides.
6. Whether approval evidence stores only a timestamp/revision or also a user-supplied note. Recommended: optional note, no chat transcript.

## 25. Official format references

- OpenAI Codex skills: https://developers.openai.com/codex/build-skills
- OpenAI Codex customization locations: https://developers.openai.com/codex/customization/overview
- Claude Code skills: https://docs.anthropic.com/en/docs/claude-code/skills
- Claude Code subagents: https://docs.anthropic.com/en/docs/claude-code/sub-agents
- GitHub Copilot agent skills: https://docs.github.com/en/copilot/concepts/agents/about-agent-skills
- GitHub Copilot customization locations: https://docs.github.com/en/copilot/reference/customization-cheat-sheet
