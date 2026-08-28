# Structured Result Schemas

Keep every result compact, evidence-based, and free of hidden reasoning.

## Discovery result

```yaml
summary: concise lens-specific conclusion
findings:
  - evidence: repository path, command, or user statement
    implication: why it matters
questions:
  - question: user decision required
    reason: why inspection cannot resolve it
    recommendation: preferred answer when justified
    consequences: material tradeoffs
assumptions:
  - statement: safely inferred fact
    confidence: high|medium|low
risks:
  - risk: material risk
    mitigation: suggested treatment
roadmap_proposals: []
readiness:
  status: ready|questions-remain|blocked
  reason: brief explanation
```

## Plan packet

Use a JSON object with `roadmap_item`, `goal`, `scope`, `decisions`, `acceptance_criteria`, `constraints`, `tasks`, `risks`, and `open_questions`. A new `roadmap_item` needs `title`, `category`, and `horizon`; an existing one uses its approved `id`. Each task has `title`, `description`, and `verification`. `goal` must be non-empty and `open_questions` must be empty before recording. Task verification is declarative; command output remains evidence rather than plan content.

## Implementation result

```yaml
task: task identifier
status: completed|blocked|amendment-required
changes:
  - path: changed path
    summary: concise change summary
tactical_variances: []
verification:
  - command: command run
    result: passed|failed|not-run
    evidence: concise result
amendment: null
```

## Verification result

```yaml
status: passed|repairable|amendment-required|blocked
criteria:
  - criterion: approved acceptance criterion
    result: passed|failed|not-verifiable
    evidence: concise evidence
issues:
  - severity: blocking|important|advisory
    finding: concise issue
    treatment: repair|amendment|roadmap-proposal|none
recommended_next_action: complete|repair|amend|block
```
