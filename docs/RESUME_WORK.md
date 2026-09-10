# Resume interrupted work

This is an operational companion to the [Cassor PRD](PRD.md)'s context and
token-efficiency principles.

Start a fresh session with compact context instead of loading a full history:

```sh
cassor context --json
cassor context --item ITEM_ID --role implementation --max-bytes 12000
```

Use the active task and its next action as the starting point. If state must move between checkouts, export it first:

```sh
cassor export --file cassor-state.json
cassor import --file cassor-state.json
```

Run `cassor check` after resuming to find blocked work or lifecycle records that need reconciliation.
