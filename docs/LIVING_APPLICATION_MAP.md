# Living Application Map

## Product direction

Cassor should grow into a living product memory for an application: a
one-stop workspace for understanding existing capabilities, defining the next
ones, approving delivery work, and retaining the evidence behind the result.

It is not only a task tracker or an agent orchestrator.  The primary user
experience is feature authoring and application understanding; agents are
contributors to that shared, curated record.

## Two kinds of truth

Each feature record should make two related things clear:

| Dimension | Question answered |
| --- | --- |
| Current product truth | What does the application do today? What is supported, limited, deprecated, or known to be missing? |
| Intended change | What should change next, why, what was decided, and how will success be verified? |

A shipped capability remains in Cassor after delivery.  For example,
Authentication retains its current behavior, constraints, decisions, and
known gaps.  A future change such as Passkeys is linked to Authentication as a
child feature or change proposal rather than replacing its history.

## Core model

The enduring unit is a **feature dossier**, not a one-time roadmap task.

```text
Application
  |- shipped capabilities
  |- in-flight features
  |- planned opportunities
  |- known gaps and technical debt
  `- cross-cutting decisions and constraints

Feature dossier
  |- current-state summary
  |- problem, intent, and scope
  |- research, risks, assumptions, and decisions
  |- acceptance criteria
  |- change proposals and approved delivery plans
  |- tasks, implementation record, and verification evidence
  `- relationships to other features and decisions
```

Existing Cassor roadmap items, plan revisions, tasks, lifecycle phase records,
acceptance criteria, and feature reports can become parts of this dossier. The
goal is an incremental evolution rather than a rewrite.

## Agent collaboration

Agents contribute compact, evidence-based artifacts to a feature dossier.
For example, a research agent may submit a research finding; an implementation
agent may add an implementation note; and a verification agent may add
verification evidence.

Each artifact should identify its kind, author role, status, summary,
supporting evidence, timestamps, and any superseded artifact.  Artifacts are
append-only or revisioned, rather than silently overwritten.

The orchestrator remains the editorial and workflow authority:

- Workers may write only artifact types within their assigned scope.
- Submitted artifacts may be visible as unreviewed, but only accepted artifacts
  may influence a plan or workflow transition.
- Only the orchestrator may approve plans, waive criteria, change canonical
  feature state, or make roadmap transitions.
- Persist compact conclusions and evidence, never raw worker transcripts or
  hidden reasoning.

An agent's `role` (for example, `research`, `implementation`, or
`verification`) describes who contributed an artifact. It must not grant that
agent ownership of the feature's shared current truth.

## Context as a projection

Agent context should be a role- and feature-scoped view of the dossier, not a
copy of the entire database and not an opaque agent memory store.  For example,
a research context can contain the current-state summary, accepted decisions,
open questions, relevant prior findings, and known risks.  An implementation
context adds the approved plan and acceptance criteria.

This keeps handoffs resumable while preventing stale, speculative, or
irrelevant notes from being treated as canonical.

## Delivery sequence

This direction is an initiative composed of independently useful features:

1. **Feature dossier foundation** — establish durable feature records with a
   current-state summary, feature classification/status, and relationships.
2. **Application map** — show shipped capabilities, in-flight work, planned
   work, and known gaps as complementary views.
3. **Feature-writing workflow** — guide authors through problem, intent,
   scope, assumptions, decisions, and acceptance criteria.
4. **Agent contribution artifacts** — persist scoped, attributed findings with
   explicit acceptance by the orchestrator.
5. **Change proposals and delivery history** — allow an enduring feature to
   receive multiple approved changes over time.
6. **Cross-feature decisions and dependencies** — expose the decisions and
   relationships that affect more than one feature.

The recommended first delivery is **Feature dossier foundation**. It creates
the durable capability record that makes the application map, guided authoring,
and agent contributions coherent.

## Capability and change boundary

Cassor keeps `Capability`, `Change`, and `Gap` as distinct feature types. A
delivery change may be linked explicitly to an enduring capability with a
typed change-of dossier link. A capability can therefore receive multiple
changes without replacing its prior delivery history.

Current product truth is stored on the capability and refreshed only through
an explicitly accepted state update sourced from a completed linked change.
Those updates are append-only, so prior accepted states remain traceable.
Discovery and other agents may record attributed dossier artifacts as `Draft`,
but drafts cannot approve plans, complete work, or become canonical product
truth until the orchestrator explicitly accepts them with evidence.

The CLI and generated roadmap expose these as complementary projections:
capabilities describe current truth, changes describe intended or delivered
work, and gaps describe known missing behavior. SQLite remains authoritative;
the context JSON and static site are derived views.

## Intended lifecycle

```text
Discover existing capability
  -> document current state
  -> identify a gap or desired change
  -> research and define it
  -> approve a delivery plan
  -> implement and verify
  -> refresh the feature's current-state summary
```

Refreshing current truth after delivery is essential: Cassor should describe
the application as it is now, while preserving the decisions and evidence that
explain how it arrived there.
