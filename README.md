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

Use `cassor --help` and its command-specific help for operational details.

## Development

Cassor requires Go 1.26 or newer.

```sh
make test
make build
bin/cassor version
make check
```

## Releases

Cassor uses GitVersion, git-cliff, and GoReleaser together. Versions are
derived from Git tags and Conventional Commit messages: `fix:` produces a
patch version, `feat:` produces a minor version, and `!` or a `BREAKING CHANGE:`
footer produces a major version. The first release is `v0.1.0`.

For local versioned builds, install the pinned tool versions:

```sh
dotnet tool install --global GitVersion.Tool --version 6.8.2
cargo install git-cliff --version 2.14.1
```

Use the selected release version for a local build, then use `bin/cassor
version` to show the version with its commit and build date. The release
workflow supplies GitVersion's calculated value automatically from the release
tag.

```sh
VERSION=v0.1.0 make build
bin/cassor version
```

Preview release notes and release-shaped artifacts without publishing them:

```sh
git-cliff --config cliff.toml --unreleased --output /tmp/release-notes.md
goreleaser check
goreleaser release --snapshot --clean --release-notes=/tmp/release-notes.md
```

The **PR Build** workflow runs golangci-lint, `gofmt`, `go vet`, tests, a Go
build, version calculation, release-note generation, and GoReleaser validation
for every pull request. Each push to `main`, including a merged pull request,
automatically starts **Release**. It repeats the Go quality gates before it
verifies the selected version is unused, generates release notes, creates the
annotated tag, and publishes the GitHub release with macOS and Linux
AMD64/ARM64 archives and checksums. Protect `main` to require pull requests if
releases must only originate from merges.

If publishing fails after the tag is created, correct the problem in a new
commit on `main` after an authorized maintainer removes the incomplete GitHub
release and tag. The resulting push reruns the workflow; it intentionally
refuses to replace either one.

## Optional documentation site

`cassor site build` generates a static snapshot under `docs/roadmap/`; it is a
derived view and is not committed. `cassor site serve` previews it locally.

The Cassor source repository also includes an enhanced SvelteKit documentation
workspace. Node 20 or newer is required only for that workspace:

```sh
cd docs-site
npm run dev
```
