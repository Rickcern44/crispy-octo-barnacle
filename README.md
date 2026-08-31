# Cassor

Cassor is a lightweight, agent-agnostic CLI for planning approved work, tracking roadmap state, and generating a static roadmap.

## Development

Cassor currently requires Go 1.26 or newer.

```sh
go test ./...
go run . --help
```

## Developer documentation site

Cassor generates the static multi-page developer site into `docs/roadmap/`.
Node 20 or newer is required only to build the documentation site.

```sh
go run . site build
go run . site serve
```

For live documentation development, use the small Astro workspace:

```sh
cd docs-site
npm run dev
```

When building for a GitHub Pages project URL, set its repository path before
running `cassor site build`, for example `SITE_BASE=/cassor go run . site build`.
Deployment automation is intentionally not configured yet.

The project contract and phased delivery plan live in [`docs/CASSOR_CODEX_HANDOFF.md`](docs/CASSOR_CODEX_HANDOFF.md).
