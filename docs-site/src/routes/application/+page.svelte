<script lang="ts">
	import { base } from '$app/paths';
	import { changesForCapability, relationshipsForItem, roadmap, stateHistoryForCapability } from '$lib/roadmap';

	const capabilities = roadmap.items.filter((item) => item.feature_type === 'Capability');
	const gapCount = roadmap.items.filter((item) => item.feature_type === 'Gap').length;
	const groups = ['Core workflow', 'Delivery and evidence', 'Context and recovery', 'Runtime and site'];
	let selectedID = capabilities[0]?.id ?? 0;
	$: selected = capabilities.find((item) => item.id === selectedID);
	$: focusedRelationships = selected ? relationshipsForItem(selected.id) : [];

	function groupFor(title: string) {
		const value = title.toLowerCase();
		if (value.includes('plan') || value.includes('approval') || value.includes('lifecycle')) return 'Core workflow';
		if (value.includes('context') || value.includes('recover') || value.includes('portable')) return 'Context and recovery';
		if (value.includes('runtime') || value.includes('site') || value.includes('roadmap')) return 'Runtime and site';
		return 'Delivery and evidence';
	}

	function linkedChanges(id: number) {
		return changesForCapability(id);
	}

	function relationshipLabel(id: number) {
		const relationships = relationshipsForItem(id);
		return relationships.length === 0 ? 'No recorded relationships' : `${relationships.length} recorded relationship${relationships.length === 1 ? '' : 's'}`;
	}
</script>

<svelte:head>
	<title>Current state · {roadmap.project_name}</title>
	<meta name="description" content="Inspect the current capability records and delivery history in local Cassor state." />
</svelte:head>

<main class="mx-auto max-w-6xl px-5 py-8 sm:px-8 lg:py-12">
	<div class="max-w-3xl">
		<p class="text-xs font-semibold tracking-[0.22em] text-cassor-400">CASSOR · CURRENT STATE</p>
		<h1 class="mt-3 text-4xl font-bold tracking-tight text-white sm:text-5xl">Recorded product state</h1>
		<p class="mt-4 text-lg leading-8 text-slate-300">This optional view shows the capability records, delivery changes, and relationships currently stored in local Cassor state. It is a derived view, not the product definition.</p>
	</div>

	{#if capabilities.length === 0}
		<section class="mt-10 rounded-2xl border border-dashed border-slate-700 bg-slate-900/50 p-8" aria-labelledby="empty-title">
			<h2 id="empty-title" class="text-xl font-bold text-white">Capability records are not populated yet</h2>
			<p class="mt-3 max-w-2xl leading-7 text-slate-300">The roadmap contains delivery changes, but no enduring capability records are currently available. The PRD explains the skill-first workflow and the intended product model.</p>
			<a class="mt-5 inline-flex rounded-lg bg-cyan-400 px-4 py-2 font-semibold text-slate-950 hover:bg-cyan-300" href={`${base}/guides/product-requirements/`}>Read the product requirements</a>
		</section>
	{:else}
		<div class="mt-10 space-y-10">
			{#each groups as group}
				{@const members = capabilities.filter((item) => groupFor(item.title) === group)}
				{#if members.length > 0}
					<section aria-labelledby={group.replaceAll(' ', '-').toLowerCase()}>
						<div class="flex items-end justify-between gap-4"><div><h2 id={group.replaceAll(' ', '-').toLowerCase()} class="text-2xl font-bold text-white">{group}</h2><p class="mt-1 text-sm text-slate-400">Enduring product capabilities and their evidence.</p></div><span class="text-sm text-slate-500">{members.length}</span></div>
						<div class="mt-4 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
							{#each members as item}
								{@const changes = linkedChanges(item.id)}
								{@const states = stateHistoryForCapability(item.id)}
								<a class="group flex min-h-56 flex-col rounded-2xl border border-slate-800 bg-slate-900/70 p-5 transition hover:-translate-y-0.5 hover:border-cyan-400 hover:bg-slate-900" href={`${base}/roadmap/rm-${item.id}/`}>
									<div class="flex items-start justify-between gap-3"><span class="rounded-full border border-emerald-400/30 bg-emerald-400/10 px-2.5 py-1 text-xs font-semibold text-emerald-200">Capability</span><span class="text-xs text-slate-500">RM-{item.id}</span></div>
									<h3 class="mt-4 text-lg font-semibold text-white group-hover:text-cyan-200">{item.title}</h3>
									<p class="mt-2 line-clamp-3 text-sm leading-6 text-slate-300">{item.current_state || item.description || 'Current behavior has not been documented yet.'}</p>
									<div class="mt-auto flex flex-wrap gap-2 pt-5 text-xs text-slate-400"><span class="rounded-full bg-slate-800 px-2.5 py-1">{changes.length} linked change{changes.length === 1 ? '' : 's'}</span><span class="rounded-full bg-slate-800 px-2.5 py-1">{gapCount} known gap{gapCount === 1 ? '' : 's'}</span><span class="rounded-full bg-slate-800 px-2.5 py-1">{states.length} accepted state{states.length === 1 ? '' : 's'}</span><span class="rounded-full bg-slate-800 px-2.5 py-1">{relationshipLabel(item.id)}</span></div>
								</a>
							{/each}
						</div>
					</section>
				{/if}
			{/each}
			{#if selected}
				<section class="rounded-2xl border border-cyan-400/30 bg-cyan-400/5 p-6" aria-labelledby="relationship-title">
					<div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between"><div><p class="text-xs font-semibold uppercase tracking-[0.18em] text-cyan-300">Focused relationship view</p><h2 id="relationship-title" class="mt-2 text-2xl font-bold text-white">How this capability connects</h2><p class="mt-1 text-sm text-slate-300">Choose a capability to inspect its recorded links and direction.</p></div><label class="text-sm text-slate-300">Focus <select bind:value={selectedID} class="ml-2 rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-slate-200">{#each capabilities as capability}<option value={capability.id}>{capability.title}</option>{/each}</select></label></div>
					<div class="mt-5 grid gap-3 sm:grid-cols-3"><a class="rounded-xl border border-cyan-400/40 bg-slate-950/60 p-4" href={`${base}/roadmap/rm-${selected.id}/`}><span class="text-xs text-cyan-300">Focused capability</span><span class="mt-2 block font-semibold text-white">{selected.title}</span></a>{#if focusedRelationships.length > 0}{#each focusedRelationships as relationship}<a class="rounded-xl border border-slate-700 bg-slate-950/40 p-4 hover:border-cyan-400" href={`${base}/roadmap/rm-${relationship.source_item_id === selected.id ? relationship.target_item_id : relationship.source_item_id}/`}><span class="text-xs uppercase tracking-wider text-slate-500">{relationship.relationship_type}</span><span class="mt-2 block font-semibold text-slate-200">{relationship.source_item_id === selected.id ? relationship.target_item_title : relationship.source_item_title}</span></a>{/each}{:else}<p class="rounded-xl border border-dashed border-slate-700 p-4 text-sm text-slate-400 sm:col-span-2">No capability relationships have been recorded yet. Linked delivery changes are available on the capability detail page.</p>{/if}</div>
				</section>
			{/if}
		</div>
	{/if}
</main>
