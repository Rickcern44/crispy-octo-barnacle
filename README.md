# Cassor

Cassor is a skill-first workflow for agent-assisted software development. It
helps a person approve concise, verifiable work while keeping the durable,
repository-local context needed to resume it later.

The product definition, current implementation boundary, and intended
direction live in [the PRD](docs/PRD.md).

## Start here

Initialize Cassor in a Git repository, then inspect its local state:

```sh
cassor init
cassor check
cassor context --json
```

Record a user-approved plan packet before implementation begins:

```sh
cassor plan record --file plan.json --approve --approved-by-user
```

See [Getting started](docs/GETTING_STARTED.md), [the workflow guide](docs/WORKFLOW_GUIDE.md), and [the context contract](docs/CASSOR_CONTEXT_CONTRACT.md) for the operational details.

## Development

Cassor requires Go 1.26 or newer.

```sh
make test
make build
bin/cassor version
make check
```

`make snapshot` writes local release-shaped artifacts to `dist/`; those files
are build output and are not committed. Publishing is intentionally not
automated.

## Optional documentation site

`cassor site build` generates a static snapshot under `docs/roadmap/`; it is a
derived view and is not committed. `cassor site serve` previews it locally.

The Cassor source repository also includes an enhanced SvelteKit documentation
workspace. Node 20 or newer is required only for that workspace:

```sh
cd docs-site
npm run dev
```
