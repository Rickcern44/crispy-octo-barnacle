# View the roadmap

The roadmap is a derived operational view; the [Cassor PRD](PRD.md) defines
the product, and `.cassor/` remains repository-specific workflow state.

Generate a fresh static projection from SQLite state:

```sh
cassor site build
```

Open `docs/roadmap/index.html` or run `cassor site serve`. A normal Cassor
project receives a self-contained roadmap with search, filters, plans, tasks,
and acceptance criteria. Repositories that include the optional `docs-site/`
workspace receive enhanced roadmap and documentation views. Rebuild this
output as needed; do not use it as a competing source of truth.

For a GitHub Pages project path, set the base path during the build:

```sh
SITE_BASE=/cassor cassor site build
```
