# RM-4 roadmap website assessment

Date: 2026-09-08
Status: Provisional findings; RM-4 implementation is tracked by approved plan revision 8.
Roadmap item: RM-4 — Build the Carbon developer roadmap dashboard.

## Question and scope

How is the roadmap website hosted, what does it look like, where can it improve,
and what value should it provide?

This was a bounded, read-only assessment of the current repository, RM-4 state,
generated site, and desktop homepage in Chrome. No implementation, deployment,
or roadmap-item status, progress, or specification changes were made during
the assessment. The assessment was tracked by approved RM-4 plan revision 7
and task 84; the resulting local dashboard work was approved and delivered
under plan revision 8 and task 85. Mobile layout, accessibility, and all
detail-page interactions were not comprehensively tested. Existing uncommitted
changes were included in the inspected snapshot and must be preserved.

## Hosting and repository evidence

The current pipeline is:

`Cassor SQLite state + Markdown docs → generated SvelteKit source → static HTML/assets + Pagefind index`

- `internal/site/site.go` exports state into
  `docs-site/src/lib/generated/roadmap.json`, creates guide routes from repository
  Markdown, and compiles the site.
- `docs-site/package.json` uses SvelteKit, Svelte, Tailwind, mdsvex, and Pagefind.
- `docs-site/svelte.config.js` uses the static adapter, outputs to `docs/roadmap/`,
  and accepts a `SITE_BASE` environment variable for a hosting subpath.
- `internal/cmd/site.go` implements `cassor site build` and `cassor site serve`.
  The default local address is `127.0.0.1:8080`; a watch option rebuilds on database
  changes.
- The generated site needs no runtime database or application server.
- No Git remote, deployment workflow, or configured public URL was found in this
  checkout. This does not establish whether a deployment exists elsewhere.
- `README.md` explicitly says deployment is not automated, but incorrectly calls
  the development workspace Astro. `docs-site/README.md` is Svelte scaffold
  text, consistent with the actual SvelteKit package configuration.

An assessment-only preview served the existing output at `http://127.0.0.1:4321/`.
That temporary address is not a deployment or a durable handoff dependency.

## Current experience and findings

The desktop homepage has a coherent dark navy palette, cyan accents, clear
typography, rounded cards, and an alternating delivery timeline. Its large
introduction and generous spacing feel closer to a delivery-history page than
RM-4's original high-density, detail-first dashboard brief. Later recorded RM-4
tasks explicitly introduced the alternating tree and simplified status display;
these are prior design choices to reconsider deliberately, not assumed defects.

| Finding | Consequence | Proposed improvement |
| --- | --- | --- |
| Large introduction and duplicate Roadmap/Docs navigation | Useful delivery information starts lower on the page | Compress the header and surface active work immediately |
| Timeline sorts oldest dated work first | History dominates current priorities | Lead with Now, Next, and Needs attention; retain history as another view |
| Homepage shows 20 shipped features but zero recorded capabilities | The application map does not explain what the product can do | Populate capability records and distinguish shipped changes from enduring capabilities |
| RM-6 and RM-23 share a title but appear as planned and active respectively | Their relationship is unclear to readers | Inspect their scope and relationships before reconciling or explaining the overlap |
| RM-4 is In Progress at 35%, with unchecked specifications and completed recorded tasks | Progress and evidence appear inconsistent | Reconcile acceptance evidence, checklist, and lifecycle state; do not infer completion from task history alone |
| Application-map groups use `slice(0, 5)` without a View all link | Larger groups silently hide entries | Add complete filtered views and meaningful count links |
| No visible generation time or source revision | Readers cannot judge snapshot freshness | Show when and from which revision the site was generated |
| Search is implemented but absent from the observed homepage | Navigation may be harder than intended | Diagnose initialization and verify search against the generated production output |

The detail-page source already includes ownership and delivery metadata,
specifications, implementation history, task outcomes, capability/change links,
relationships, accepted state history, dossier findings, acceptance criteria,
and orchestration reports. These are valuable building blocks for a trustworthy
overview rather than reasons to build a separate tracking system.

Relevant presentation sources:

- `docs-site/src/routes/+page.svelte`
- `docs-site/src/routes/+layout.svelte`
- `docs-site/src/routes/layout.css`
- `docs-site/src/routes/roadmap/[id]/+page.svelte`
- `docs-site/src/lib/roadmap.ts`
- `docs-site/src/lib/Search.svelte`

## Intended value

The website should give a human a trustworthy view of agent-driven work:

1. Understand the product: what can Cassor do today?
2. Direct the work: what is underway, blocked, or awaiting a decision?
3. Verify delivery: what changed, why, and what evidence supports completion?
4. Resume work: where should someone pick up without reconstructing conversations?

Success should be judged by how quickly a reader can answer these questions and
reach supporting evidence, rather than by the volume of displayed history or the
number of shipped items alone.

## Options and recommendation

### Presentation

- Polish the existing timeline: smallest change, but leaves current priorities and
  decision support secondary.
- Make a developer overview the primary page: recommended. Put active work and
  decisions first, capabilities second, and delivery history third. Preserve
  detailed evidence and the existing timeline through clear navigation.
- Build an interactive editing application: adds runtime and synchronization
  responsibilities without an established need in this assessment.

Keep the static architecture for this pass. Static delivery remains compatible
with client-side filtering and search. Its tradeoff is snapshot freshness:
readers only see state included in the latest successful build and publication.

### Hosting

Local preview is sufficient for personal repository work. If a shareable public
site is desired, GitHub Pages is a reasonable fit: build the static output and
publish `docs/roadmap/` as an artifact, with the correct base path. GitHub supports
[custom static-site deployment workflows](https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages).

No hosting migration is justified by the current evidence. Before implementing
publication, establish the destination, intended audience, and how the build
receives the canonical roadmap snapshot. Decide which internal evidence belongs
in the published artifact. A hosting choice or publication is not approved by
this spike.

## Open decisions and Cassor continuation

- Is the primary reader Rick working locally, contributors, or public users?
- Should the primary view emphasize active delivery, the product capability map,
  or both? The recommendation is active delivery first.
- Which decisions and evidence are useful on the overview versus detail pages?
- Should this pass include deployment, or only improve the local experience?
- What explains the RM-6/RM-23 overlap and RM-4's progress/checklist state?

Resume through the Cassor workflow using this file as discovery evidence. Read
the current applicable skill and scoped RM-4 context; do not treat this artifact
as canonical state. The current checkout records the assessment under RM-4 plan
revision 7 and the approved local dashboard implementation under plan revision
8. The roadmap item remains In Progress at 35%; its existing specifications
remain unchanged. Recheck observations before further implementation because
the assessed worktree contained ongoing changes.

Proposed implementation boundaries for planning:

1. Reconcile the relevant roadmap records using supported workflow operations.
2. Define the overview's information hierarchy and acceptance criteria.
3. Implement the compact overview, complete navigation/filtering, and freshness
   display; investigate the missing search UI.
4. Verify desktop/mobile behavior, evidence navigation, search, and hosted-subpath
   links with checks appropriate to the approved changes.
5. Correct the site documentation; handle publishing only if included in scope.

This spike informed the focused local dashboard continuation in approved RM-4
plan revision 8. It does not modify the SDD Lite Improvement Plan or change
roadmap item status.
