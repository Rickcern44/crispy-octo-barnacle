# View the roadmap

Generate a fresh static projection from SQLite state:

```sh
cassor site build
```

Open `docs/roadmap/index.html` or run `cassor site serve`. A normal Cassor
project receives a self-contained roadmap with search, filters, plans, tasks,
and acceptance criteria. Repositories that include the optional `docs-site/`
workspace receive Cassor's enhanced Application, Roadmap, and Docs views.

For a GitHub Pages project path, set the base path during the build:

```sh
SITE_BASE=/cassor cassor site build
```
