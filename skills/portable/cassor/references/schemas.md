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

Use a JSON object with `roadmap_item`, `goal`, `scope`, `decisions`, `acceptance_criteria`, `constraints`, `tasks`, `risks`, and `open_questions`. For an existing item, preserve its approved `id` and roadmap fields:

```json
{
  "roadmap_item": {"id": 42, "title": "Example item", "description": "What the item delivers", "category": "Needed", "horizon": "Now", "rationale": "Why it matters"},
  "goal": "Implement the example item",
  "scope": {"included": ["Approved implementation work"], "excluded": ["Unrelated roadmap work"]},
  "decisions": ["Use the existing project conventions"],
  "acceptance_criteria": [{"id": "C1", "title": "The item is implemented", "description": "The outcome is complete and verified.", "required": true}],
  "constraints": ["Preserve existing behavior outside this item"],
  "tasks": [{"title": "Implement the item", "description": "Complete the implementation.", "verification": ["go test ./..."]}],
  "risks": ["Implementation assumptions may need refinement"],
  "open_questions": []
}
```

Generate a starting packet with `cassor plan template --item ID > plan.json`, edit the packet, and run `cassor plan validate --file plan.json`. Only after explicit user approval, record it with `cassor plan record --file plan.json --approve --approved-by-user`. `goal` must be non-empty and `open_questions` must be empty before recording. Task verification is declarative; command output remains evidence rather than plan content.

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
