<script lang="ts">
	import { base } from '$app/paths';
	import { milestoneDate, milestoneKind, milestones, roadmap, shippedFeatureCount } from '$lib/roadmap';

	const scheduled = milestones.filter((item) => milestoneKind(item) !== 'unscheduled');
	const unscheduled = milestones.filter((item) => milestoneKind(item) === 'unscheduled');
</script>

<svelte:head>
	<title>{roadmap.project_name} delivery roadmap</title>
	<meta name="description" content="Verified work and planned delivery milestones from Cassor." />
</svelte:head>

<main class="mx-auto max-w-6xl px-5 py-12 sm:px-8 lg:py-20">
	<header class="max-w-3xl">
		<p class="text-sm font-semibold tracking-[0.22em] text-cassor-400">CASSOR · DELIVERY ROADMAP</p>
		<h1 class="mt-3 text-4xl font-bold tracking-tight text-white sm:text-6xl">{roadmap.project_name}</h1>
		<p class="mt-5 text-lg leading-8 text-slate-300">A date-based view of delivered work and the planned features that move this repository forward.</p>
		<nav class="mt-7 flex flex-wrap gap-3" aria-label="Documentation navigation">
			<a class="rounded-full bg-cyan-400 px-4 py-2 text-sm font-semibold text-slate-950 transition hover:bg-cyan-300" href={`${base}/`}>Roadmap</a>
			<a class="rounded-full border border-slate-600 px-4 py-2 text-sm font-semibold text-slate-200 transition hover:border-cyan-400 hover:text-cyan-300" href={`${base}/guides/project-handoff/`}>Docs</a>
		</nav>
		<p class="mt-6 inline-flex rounded-full border border-emerald-400/30 bg-emerald-400/10 px-4 py-2 text-sm font-semibold text-emerald-300">🏁 {shippedFeatureCount} features shipped</p>
	</header>

	<section class="mt-14" aria-labelledby="timeline-title">
		<div class="flex items-center justify-between gap-4">
			<h2 id="timeline-title" class="text-2xl font-bold text-white">Delivery timeline</h2>
			<p class="text-sm text-slate-400">{scheduled.length} dated milestones</p>
		</div>
		<div class="relative mt-8 border-l border-slate-700 pl-7 sm:pl-10 lg:border-l-0 lg:pl-0">
			{#each scheduled as item, index}
				<div class={`relative mb-4 lg:grid lg:grid-cols-[1fr_3rem_1fr] lg:items-center`}>
					<span class={`absolute -left-[2.15rem] top-5 h-4 w-4 rounded-full border-4 border-cassor-950 ${milestoneKind(item) === 'done' ? 'bg-emerald-400' : 'bg-cyan-400'} lg:static lg:col-start-2 lg:row-start-1 lg:mx-auto`}></span>
					<span class="absolute left-1/2 top-0 hidden h-full w-px -translate-x-1/2 bg-slate-700 lg:block"></span>
					<a class={`group relative block rounded-xl border border-slate-800 bg-slate-900/70 p-4 shadow-sm transition hover:-translate-y-0.5 hover:border-cyan-400 hover:bg-slate-900 lg:row-start-1 ${index % 2 === 0 ? 'lg:col-start-1 lg:text-right' : 'lg:col-start-3'}`} href={`${base}/roadmap/rm-${item.id}/`}>
					<div class="flex flex-wrap items-center gap-x-4 gap-y-2">
						<p class="text-sm font-semibold tracking-wider text-cyan-300">{milestoneDate(item)}</p>
						<span class="text-xs font-medium uppercase tracking-wider text-slate-400">{milestoneKind(item) === 'done' ? 'Delivered' : 'Planned'} · RM-{item.id}</span>
					</div>
					<h3 class="mt-2 text-xl font-semibold text-white group-hover:text-cyan-200">{item.title}</h3>
					<p class={`mt-3 text-sm font-semibold ${milestoneKind(item) === 'done' ? 'text-emerald-300' : 'text-cyan-300'}`}>{item.status}</p>
				</a>
				</div>
			{/each}
		</div>
	</section>

	{#if unscheduled.length > 0}
		<section class="mt-14" aria-labelledby="unscheduled-title">
			<h2 id="unscheduled-title" class="text-2xl font-bold text-white">Unscheduled</h2>
			<div class="mt-6 grid gap-4 sm:grid-cols-2">
				{#each unscheduled as item}
					<a class="rounded-2xl border border-slate-800 bg-slate-900/60 p-5 transition hover:border-cyan-400" href={`${base}/roadmap/rm-${item.id}/`}><p class="text-xs font-semibold tracking-wider text-slate-400">RM-{item.id} · {item.status}</p><h3 class="mt-2 font-semibold text-white">{item.title}</h3></a>
				{/each}
			</div>
		</section>
	{/if}
</main>
