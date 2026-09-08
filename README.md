# Cassor

Cassor is a lightweight, agent-agnostic CLI for planning approved work, tracking roadmap state, and generating a static roadmap.

## Development

Cassor currently requires Go 1.26 or newer.

```sh
go test ./...
go run . --help
```

## Local builds and versions

Use the repeatable local targets:

```sh
make test
make build
bin/cassor version
make check
```

The binary reports its Git-derived version, commit, and UTC build date. Untagged
builds use the current Git description; a `-dirty` suffix means local changes
were included. Release versions are created by tagging a clean, verified commit
with SemVer, for example `git tag -a v0.1.0 -m "v0.1.0"`, then running
`make build`.

For release-shaped local artifacts, install GoReleaser and run `make release-check`
followed by `make snapshot`. Snapshot artifacts and checksums are written to
`dist/`; tagged builds use the annotated `vX.Y.Z` tag. Publishing is intentionally
not automated yet.

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
