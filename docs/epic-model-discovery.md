# Epic model discovery

## Decision

Cassor will add an **Epic** as an optional, first-class container for
Features. A Feature has zero or one Epic parent; an Epic can contain many
Features. Existing roadmap items remain valid Features without a parent.

Containment is distinct from the existing `extends`, `depends_on`, and
`replaces` feature relationships. It is also distinct from capability/change
links: those describe delivery and dossier semantics, while an Epic expresses
product grouping.

The first delivery will model Epics in a dedicated `epics` table and add a
nullable `epic_id` foreign key to `roadmap_items`. This preserves the current
`feature_type` contract (`Capability`, `Change`, or `Gap`) for Features and
avoids making a grouping concept appear in type-specific dossier projections.

```text
Epic (optional)
└── Feature (existing roadmap item)
    └── Plan revision
        └── Task
```

An Epic is a grouping record in the first delivery. It does not own plans or
tasks, and it has no automatic completion rollup. Progress reporting may show
the states of its child Features, but must not infer that the Epic itself is
complete. Moving a Feature between Epics is an explicit update and is allowed;
the current system has no historical parent-assignment record, so it must not
claim to preserve one.

## User experience

The future command surface should retain the item-centred workflow while
making the hierarchy visible:

- `cassor epic add`, `list`, `show`, and `update` manage grouping records.
- `cassor item add --epic ID` creates a Feature in an Epic.
- `cassor item update ID --epic ID` assigns or reparents a Feature; a
  `--clear-epic` form makes it ungrouped.
- Item list/show JSON includes `epic_id` and a compact Epic reference. Human
  output shows the parent when present.
- Context includes a compact parent reference and, when an Epic is scoped, a
  bounded child summary with stable commands for omitted detail.

Ungrouped Features remain a normal, supported path for small requests. A Task
continues to belong only to an approved plan revision of a Feature.

## Persistence and compatibility

The delivery migration must create `epics` and add a nullable
`roadmap_items.epic_id` reference with an index for child queries. It must
preserve all existing roadmap item IDs, plans, tasks, metadata, and feature
types. Existing databases migrate with `epic_id = NULL`; fresh databases have
the same schema after all migrations run.

Store APIs must reject a missing Epic parent and prevent invalid direct SQL
states from silently succeeding. Deleting an Epic with children must be
rejected until those Features are explicitly moved or ungrouped. The database
foreign key protects referential integrity; `cassor check` additionally
reports hierarchy violations.

The portable format must add an optional Epic record collection and nullable
`epic_id` on exported items. Imports created before Epic support, which omit
these fields, must continue to import as ungrouped Features. New imports must
validate duplicate IDs, missing parent Epics, deterministic ordering, and
round-trip preservation. The format version changes only if additive optional
fields cannot retain the documented compatibility contract.

Plan packets for new Features should accept an optional `epic_id`. Old packets
must remain valid and create ungrouped Features. The strict packet decoder and
portable importer need explicit compatibility tests for both forms.

## Delivery boundaries

The implementation plan must cover these phases in order:

1. Add migrations, store types and APIs, parent validation, portable
   import/export, and `cassor check` coverage.
2. Add Epic and parent-assignment commands, then project compact hierarchy
   references through item output and context without breaking byte limits.
3. Add hierarchy data to the generated documentation site and the Svelte site,
   then update the PRD, README, CLI help, and portable Cassor skill.

Verification must include fresh and upgrade migration tests, store-invariant
tests, portable legacy and round-trip tests, packet compatibility tests, CLI
tests, bounded-context tests, generated-site tests, `go test ./...`, and
`cassor check`.

This discovery record intentionally makes no schema, command, or runtime
behaviour change. Those changes require a separately approved implementation
plan.
