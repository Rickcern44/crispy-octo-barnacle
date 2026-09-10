<script lang="ts">
	import { base } from '$app/paths';
	import { roadmap } from '$lib/roadmap';

	const guide = { label: 'Product requirements', slug: 'product-requirements', description: 'Understand Cassor’s skill-first identity, principles, current boundaries, and intended direction.' };
</script>

<svelte:head><title>Documentation · {roadmap.project_name}</title><meta name="description" content="Task-oriented guides and reference documentation for Cassor." /></svelte:head>

<main class="mx-auto max-w-6xl px-5 py-8 sm:px-8 lg:py-12">
	<div class="max-w-3xl"><p class="text-xs font-semibold tracking-[0.22em] text-cassor-400">CASSOR · PRODUCT</p><h1 class="mt-3 text-4xl font-bold tracking-tight text-white sm:text-5xl">One source of truth</h1><p class="mt-4 text-lg leading-8 text-slate-300">The PRD defines Cassor’s product identity, current boundary, and intended direction. Use CLI help for operational command details.</p></div>
	<a class="mt-10 block max-w-xl rounded-xl border border-slate-800 bg-slate-900/70 p-5 transition hover:border-cyan-400 hover:bg-slate-900" href={`${base}/guides/${guide.slug}/`}><h2 class="font-semibold text-cyan-200">{guide.label}</h2><p class="mt-1 text-sm leading-6 text-slate-400">{guide.description}</p></a>
	<div class="mt-12 rounded-xl border border-slate-800 bg-slate-900/50 p-5 text-sm text-slate-300">The roadmap is a derived view of current local workflow state. <a class="font-semibold text-cyan-300 hover:text-cyan-200" href={`${base}/`}>Open the roadmap</a>.</div>
</main>
