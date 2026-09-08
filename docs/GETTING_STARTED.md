# Getting started

Initialize Cassor from the root of an existing Git repository:

```sh
cassor init
cassor check
```

Cassor stores authoritative state in `.cassor/`. Build the static site whenever you want a shareable snapshot:

```sh
cassor site build
cassor site serve
```

The site is written to `docs/roadmap/` and does not require a running database or application server.
