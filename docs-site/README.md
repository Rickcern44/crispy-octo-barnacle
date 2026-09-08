# Cassor developer site

This SvelteKit workspace builds the static Cassor application map, roadmap, and documentation site. Cassor state remains authoritative; generated roadmap data and guide routes are projections.

From the repository root:

```sh
go run . site build
```

For live UI work:

```sh
cd docs-site
npm install
npm run dev
```

The production build writes to `docs/roadmap/` and creates the Pagefind search index. Set `SITE_BASE=/cassor` when the site is hosted below a repository path.

Useful checks are `npm run check`, `npm run lint`, `npm run build`, and `npm run ci`.
