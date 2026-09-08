# Cassor Plan Packet Schema

`cassor plan record --file packet.json --approve --approved-by-user` atomically records a user-approved plan revision and its pending tasks.

The packet is JSON. It must have a non-empty `goal`, no `open_questions`, and at least one task. A new roadmap item requires `title`, `category`, and `horizon`; alternatively, set `roadmap_item.id` to an existing approved item.

```json
{
  "roadmap_item": {
    "title": "Deliver the skills milestone",
    "description": "Portable skill and Codex installer.",
    "category": "Needed",
    "horizon": "Now",
    "rationale": "Required for local skill use"
  },
  "goal": "A Codex project skill installs safely and follows Cassor approval rules.",
  "scope": {
    "included": ["portable skill", "Codex project installer"],
    "excluded": ["Claude and Copilot adapters"]
  },
  "decisions": ["Project scope is the default."],
  "acceptance_criteria": [
    {"id": "install-safe", "title": "The installer is idempotent."}
  ],
  "constraints": ["No overwrite without explicit confirmation."],
  "tasks": [
    {
      "title": "Create the portable skill",
      "description": "Add the canonical protocol and role references.",
      "verification": ["go test ./..."]
    }
  ],
  "risks": [],
  "open_questions": []
}
```

Each criterion needs a stable `id` when an amendment must carry it forward; omitted IDs receive deterministic `C1`, `C2`, and so on. Criteria and task verification requirements are recorded in the same transaction as the approved plan. The optional `--approval-note` stores a brief approval record. Cassor stores the packet as immutable JSON but never stores conversation transcripts.
