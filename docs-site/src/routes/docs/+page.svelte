<script lang="ts">
	import { base } from '$app/paths';
	import { roadmap } from '$lib/roadmap';

	const groups = [
		{ title: 'Start here', links: [{ label: 'Product requirements', slug: 'product-requirements', description: 'Understand Cassor’s skill-first identity, principles, and current boundaries.' }, { label: 'Initialize a project', slug: 'getting-started', description: 'Set up repository-local Cassor state.' }, { label: 'Plan and approve work', slug: 'workflow', description: 'Record concise approved revisions and verify tasks.' }] },
		{ title: 'Work efficiently', links: [{ label: 'Resume interrupted work', slug: 'resume-work', description: 'Recover compact context and continue safely.' }, { label: 'Context contract', slug: 'context-contract', description: 'Use bounded, role-scoped context safely.' }, { label: 'Plan packet schema', slug: 'plan-packet-schema', description: 'Understand the JSON contract for approved plans.' }] },
		{ title: 'Reference', links: [{ label: 'Recovery and portability', slug: 'recovery', description: 'Back up, transfer, and recover Cassor state.' }, { label: 'Migration policy', slug: 'migration-policy', description: 'Understand legacy delivery-contract handling.' }, { label: 'View the roadmap', slug: 'roadmap-guide', description: 'Build a derived snapshot of workflow state.' }] },
	];
</script>

<svelte:head><title>Documentation · {roadmap.project_name}</title><meta name="description" content="Task-oriented guides and reference documentation for Cassor." /></svelte:head>

<main class="mx-auto max-w-6xl px-5 py-8 sm:px-8 lg:py-12">
	<div class="max-w-3xl"><p class="text-xs font-semibold tracking-[0.22em] text-cassor-400">CASSOR · DOCUMENTATION</p><h1 class="mt-3 text-4xl font-bold tracking-tight text-white sm:text-5xl">Build with confidence</h1><p class="mt-4 text-lg leading-8 text-slate-300">Find the guide for the task in front of you, then use the reference pages to understand the underlying contracts.</p></div>
	<div class="mt-10 grid gap-8 md:grid-cols-2 lg:grid-cols-3">
		{#each groups as group}
			<section aria-labelledby={group.title.replaceAll(' ', '-').toLowerCase()}><h2 id={group.title.replaceAll(' ', '-').toLowerCase()} class="text-xl font-bold text-white">{group.title}</h2><div class="mt-3 space-y-3">{#each group.links as link}<a class="block rounded-xl border border-slate-800 bg-slate-900/70 p-4 transition hover:border-cyan-400 hover:bg-slate-900" href={`${base}/guides/${link.slug}/`}><h3 class="font-semibold text-cyan-200">{link.label}</h3><p class="mt-1 text-sm leading-6 text-slate-400">{link.description}</p></a>{/each}</div></section>
		{/each}
	</div>
	<div class="mt-12 rounded-xl border border-slate-800 bg-slate-900/50 p-5 text-sm text-slate-300">The PRD defines product intent; the roadmap is a derived view of current local workflow state. <a class="font-semibold text-cyan-300 hover:text-cyan-200" href={`${base}/`}>Open the roadmap</a>.</div>
</main>
