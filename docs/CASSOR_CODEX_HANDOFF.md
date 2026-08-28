# Cassor: Codex Handoff

Use this document to continue the Cassor project in Codex on Rick's local machine.

## Startup prompt

Paste this into Codex from the directory where the new repository should be created:

> Build Cassor from the approved project contract and MVP plan in `CASSOR_CODEX_HANDOFF.md`. First inspect the local environment and restate the concrete implementation plan, including any necessary amendments. Do not create or modify repository files until I approve that plan revision. After approval, implement tasks sequentially, verify each task, and continue automatically unless a material amendment is required.

## Product identity

- Name: Cassor
- Name reference: a subtle combination/reference to Cassian Andor
- CLI executable: `cassor`
- Go module: `github.com/rickcern44/cassor`
- Project state directory: `.cassor/`
- SQLite database: `.cassor/cassor.db`
- Configuration: `.cassor/config.toml`
- Generated static roadmap: `docs/roadmap/index.html`
- Initial visibility: private while validating the product, with a possible future open-source release

## Product goal

Cassor is a lightweight, agent-agnostic alternative to heavyweight specification-driven development frameworks. It coordinates planning, explicit user approval, implementation, verification, roadmap maintenance, and resumable task state without producing excessive documentation or loading full planning transcripts into every agent context.

The eventual interaction model is:

1. Rick interacts only with a frontier-model orchestrator.
2. The orchestrator selects focused, economical planning agents.
3. Planning agents inspect the repository or greenfield problem through specific lenses.
4. The orchestrator asks Rick meaningful questions incrementally.
5. The orchestrator produces a compact, revisioned implementation plan.
6. Rick approves the plan.
7. An economical coding agent implements approved tasks.
8. Each task is independently verified before completion.
9. Material deviations return to the orchestrator for a revised plan and new approval.

The first release builds the Go CLI and static site generator. Agent skills and installers follow once the CLI contract is stable.

## Workflow invariants

### Approval

- No repository mutation begins until Rick approves a concrete plan revision.
- Approval is plan-level, not operation-level or exact-diff-level.
- One approval may authorize multiple tasks.
- Approved tasks execute automatically and sequentially with concise progress updates.
- Each task must be verified before being marked complete.
- Approval applies only to the exact plan revision that was approved.
- Approved plan revisions are immutable.

### Bounded implementation adaptation

The implementer may make tactical adjustments that preserve the approved behavior and boundaries, including following repository conventions, correcting an inaccurate file list, choosing an equivalent private helper, or adding a narrowly relevant test.

The implementer must stop and request a plan amendment before changing:

- User-visible behavior
- Scope
- Architecture or component boundaries
- Public APIs or persisted schemas
- Dependencies
- Acceptance criteria
- Compatibility guarantees
- Migration or destructive behavior
- Approved delivery outcomes

New adjacent features and opportunistic cleanup must not be silently included.

### Discovery and questions

- Planning is coverage-driven, not limited to a fixed question count.
- Planning agents inspect discoverable repository facts before asking Rick.
- Questions are asked incrementally as meaningful uncertainty appears.
- Questions should represent decisions rather than facts an agent can discover.
- Questions should include a recommendation and material tradeoffs when applicable.
- Before final plan approval, every relevant area must be answered, safely inferred, or marked irrelevant.
- The orchestrator should explicitly ask whether intended behavior remains uncovered before presenting the final plan.

Relevant coverage areas include outcome, users and permissions, current behavior, scope, edge cases, compatibility, UX/API contract, failure behavior, testing, migration, and delivery.

### Recommendations and scope

Planning agents may propose work beyond the initial request, but it cannot enter the active plan or persistent roadmap without Rick's approval.

Roadmap proposals have independent attributes:

- Category: configurable examples include Needed, Recommended, Stretch goal, Technical debt, Security, and Research.
- Horizon: initially Now, Next, or Later.
- Status: Proposed, Approved, Active, Blocked, Completed, Declined, or Archived as appropriate.

Categories must be database records rather than hard-coded enums. A category does not authorize implementation or determine execution order.

### Context control

- SQLite is the authoritative state store.
- Agents mutate state only through Cassor commands, not arbitrary SQL.
- Static HTML, Markdown, and JSON are generated projections, never competing sources of truth.
- Worker agents return compact structured findings; full worker transcripts are not persisted as shared context.
- `cassor context --json` will provide the compact context needed by future agent skills.

## Core hierarchy

```text
Roadmap item
└── Plan revision
    └── Task
```

- A roadmap item describes an outcome or opportunity.
- A plan is a revisioned, approvable implementation proposal for a roadmap item.
- Tasks are executable units tied to one exact plan revision.
- Small roadmap items may contain one plan with one task.
- Tasks cannot start until their plan revision is approved.

## Technology decisions

- Language: Go
- CLI framework: Cobra
- SQLite access: standard `database/sql`
- SQLite driver: prefer `modernc.org/sqlite` to avoid CGO
- Templates: standard `html/template`
- Embedded migrations and site assets: standard `embed`
- Preview server: standard `net/http`
- Configuration: TOML
- Distribution goal: a single cross-platform executable with no Node dependency
- Static-site output: one self-contained HTML file for the first release
- SQLite database is committed initially because this is a one-roadmap-per-repository, initially single-user workflow

## MVP Plan Revision 1

This plan was presented but had not yet been explicitly approved when the work moved to local Codex. Codex must restate it after inspecting the local environment and obtain approval before changing files.

### Task 1: Bootstrap the Go CLI

- Initialize `github.com/rickcern44/cassor`.
- Add the Cobra root command.
- Establish a pragmatic package structure.
- Add `.gitignore`, an initial README, and tests.
- Initialize local Git if the target is not already a repository.
- Do not create a GitHub repository or remote without separate authorization.

### Task 2: Repository initialization and discovery

Implement:

```bash
cassor init --name "Project Name"
```

It must:

- Locate the Git repository root.
- Create `.cassor/config.toml` and `.cassor/cassor.db`.
- Apply embedded SQLite migrations.
- Insert default categories.
- Refuse to overwrite an existing Cassor project.
- Allow subsequent commands to run from nested directories by searching upward for the Cassor configuration.

### Task 3: Initial data model

Support:

- Configurable categories
- Roadmap items with category, horizon, status, rationale, and timestamps
- Revisioned plans with draft and approved states
- Immutable approved plan revisions
- Tasks tied to one exact plan revision
- Task status, completion outcome, and lifecycle timestamps
- Referential integrity and transactional state transitions

Initial categories:

- Needed
- Recommended
- Stretch goal
- Technical debt
- Security
- Research

Initial horizons:

- Now
- Next
- Later

### Task 4: Roadmap commands

Target command surface:

```bash
cassor category add|list
cassor item add|list|show|update|approve|decline
cassor plan create|show|revise|approve
cassor task add|list|show|start|complete|block
```

Required rules:

- New roadmap items begin as proposed.
- Only approved roadmap items may receive executable plans.
- Revising a plan creates a new draft revision.
- Approved revisions cannot be modified.
- Tasks cannot start without an approved plan revision.
- Invalid transitions fail without partially modifying the database.
- Read commands support stable `--json` output.

### Task 5: Compact agent context

Implement:

```bash
cassor context
cassor context --json
```

It returns only active work, proposed items, plans awaiting approval, blocked tasks, and recently completed tasks.

### Task 6: Static roadmap site

Implement:

```bash
cassor site build
```

Generate `docs/roadmap/index.html` as a self-contained static page with:

- Project summary
- Category, horizon, and status filtering
- Active work
- Proposed roadmap items
- Plan approval state
- Completed-task history
- Responsive desktop/mobile layout
- Light and dark themes
- Embedded CSS, JavaScript, and serialized roadmap data

The build should be deterministic except for any explicitly displayed generation timestamp.

### Task 7: Preview and validation

Implement:

```bash
cassor site serve
cassor check
```

`site serve` serves the generated static site locally. `check` validates configuration, schema version, relationships, workflow invariants, approval requirements, and successful site generation.

### Task 8: Testing and self-hosting

Test repository discovery, migrations, transitions, plan immutability, JSON output, and deterministic site generation.

After the CLI is working:

- Initialize Cassor inside its own repository.
- Import or recreate the Cassor development roadmap in its SQLite database.
- Generate `docs/roadmap/index.html`.
- Use Cassor for subsequent development.

## Explicitly deferred

- Agent skill creation
- `cassor skills install`
- Codex, Claude Code, and GitHub Copilot adapters
- Model routing configuration
- GitHub repository creation and publishing
- Release automation and distributed binaries
- Dynamic editing from the roadmap webpage
- Multiple roadmaps per repository
- Multi-repository aggregation
- Multi-writer SQLite synchronization
- Formal licensing decision

## Future skills milestone

After the CLI contract stabilizes, Cassor should embed a portable workflow protocol and runtime-specific adapters. Planned commands include:

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

The installer must detect existing files, support `--dry-run`, preserve customizations, avoid unapproved overwrites, and keep the portable protocol separate from runtime-specific adapters.

## Local Codex discovery checklist

Before presenting the local implementation plan, Codex should determine:

1. Installed Go version and intended minimum supported version.
2. Whether the target directory is empty and whether Git is initialized.
3. Preferred test/assertion conventions, if the repository already establishes them.
4. Whether Rick wants an initial license now or wants it deferred.
5. Whether generated `docs/roadmap/index.html` should be committed automatically or only built on demand. The current recommendation is to commit it.
6. Whether the first implementation should include the entire MVP above or split approval into smaller milestones. The current recommendation is one approved MVP plan executed task-by-task.

## Completion standard

Cassor's initial CLI/generator milestone is complete only when:

- The repository builds with the chosen supported Go version.
- Automated tests pass.
- `cassor init` initializes an independent test repository.
- Roadmap items, plan revisions, and tasks obey approval transitions.
- `cassor context --json` emits stable machine-readable data.
- `cassor site build` generates a functional self-contained roadmap page.
- `cassor check` succeeds on valid state and explains invalid state.
- Cassor initializes and generates the roadmap for its own repository.

