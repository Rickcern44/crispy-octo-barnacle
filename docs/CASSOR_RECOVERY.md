# Cassor recovery and checkout ownership

SQLite in `.cassor/cassor.db` is the authoritative state for one checkout.
Generated files under `docs/roadmap` are a view and can be rebuilt with
`cassor site build`. A branch or worktree does not automatically share state
with another checkout; initialize and migrate each checkout that should have a
local Cassor database.

Create a deterministic backup from the repository root:

```text
cassor export --file /secure/path/cassor-state.json
```

To recover, create or use a fresh initialized checkout, apply migrations, then
import the state:

```text
cassor init --name "Cassor project"
cassor migrate
cassor import --file /secure/path/cassor-state.json
cassor check
cassor site build
```

The version 1 export includes roadmap items, categories, plan revisions and
approval state, tasks and verification requirements, lifecycle records,
criteria, criterion evidence history, task events, feature relationships, and
feature reports. It excludes schema metadata, generated roadmap files, local
configuration, transcripts, and hidden reasoning.

Imports reject unknown format versions, malformed JSON, missing references, and
non-empty logical destinations. The category rows created by `cassor init` may
remain when their IDs and names match the export. All other records are
inserted in one transaction; a validation or constraint failure rolls the
destination back without a partial restore. Version 1 deliberately has no
merge behavior, so export first when transferring state between populated
checkouts.
