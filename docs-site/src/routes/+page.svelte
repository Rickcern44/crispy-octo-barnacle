<script lang="ts">
	import { onMount } from 'svelte';
	import { base } from '$app/paths';
	import { milestoneDate, milestoneKind, milestones, roadmap, shippedFeatureCount } from '$lib/roadmap';

	type Filter = 'active' | 'planned' | 'capabilities' | 'history' | 'all';
	type Attention = { itemID: number; label: string; detail: string; href: string; tone: string };

	const planItemIDs = new Map(roadmap.plans.map((plan) => [plan.id, plan.item_id]));
	const itemByID = new Map(roadmap.items.map((item) => [item.id, item]));
	const activeItems = roadmap.items.filter((item) => ['In Progress', 'Blocked'].includes(item.status));
	const plannedItems = roadmap.items.filter((item) => ['Planned', 'Ready', 'Proposed', 'Approved'].includes(item.status));
	const capabilityItems = roadmap.items.filter((item) => item.feature_type === 'Capability');
	const historyItems = roadmap.items.filter((item) => item.status === 'Done');
	const attention: Attention[] = [];

	for (const task of roadmap.tasks.filter((task) => task.status === 'Blocked')) {
		const itemID = planItemIDs.get(task.plan_id);
		if (itemID) attention.push({ itemID, label: 'Blocked task', detail: task.title, href: `${base}/roadmap/rm-${itemID}/`, tone: 'border-rose-400/30 bg-rose-400/10 text-rose-200' });
	}
	for (const plan of roadmap.plans.filter((plan) => plan.status === 'Draft')) {
		attention.push({ itemID: plan.item_id, label: 'Plan awaiting approval', detail: plan.item_title, href: `${base}/roadmap/rm-${plan.item_id}/`, tone: 'border-amber-400/30 bg-amber-400/10 text-amber-200' });
	}
	for (const criterion of roadmap.criteria.filter((criterion) => criterion.Status === 'Failed')) {
		attention.push({ itemID: criterion.ItemID, label: 'Failed verification', detail: criterion.Title, href: `${base}/roadmap/rm-${criterion.ItemID}/`, tone: 'border-rose-400/30 bg-rose-400/10 text-rose-200' });
	}
	const reconciliationCounts = new Map<number, number>();
	for (const task of roadmap.completed_tasks) {
		const itemID = planItemIDs.get(task.plan_id);
		const item = itemID ? itemByID.get(itemID) : undefined;
		if (itemID && item && item.status !== 'Done') reconciliationCounts.set(itemID, (reconciliationCounts.get(itemID) ?? 0) + 1);
	}
	for (const [itemID, count] of reconciliationCounts) {
		attention.push({ itemID, label: 'Completed tasks need reconciliation', detail: `${count} completed task${count === 1 ? '' : 's'} belong to an item still in progress`, href: `${base}/roadmap/rm-${itemID}/`, tone: 'border-cyan-400/30 bg-cyan-400/10 text-cyan-200' });
	}

	const scheduled = milestones.filter((item) => milestoneKind(item) !== 'unscheduled').reverse();
	const unscheduled = milestones.filter((item) => milestoneKind(item) === 'unscheduled');
	const generatedLabel = roadmap.generated_at.replace('T', ' ').replace('Z', ' UTC');
	let selectedFilter: Filter = 'active';
	let search = '';
	let selectedHorizon = 'all';
	let selectedStatus = 'all';
	onMount(() => {
		const params = new URLSearchParams(window.location.search);
		selectedFilter = (params.get('view') as Filter) || 'active';
		search = params.get('q') || '';
		selectedHorizon = params.get('horizon') || 'all';
		selectedStatus = params.get('status') || 'all';
	});
	function updateURL() {
		const params = new URLSearchParams();
		if (selectedFilter !== 'active') params.set('view', selectedFilter);
		if (search) params.set('q', search);
		if (selectedHorizon !== 'all') params.set('horizon', selectedHorizon);
		if (selectedStatus !== 'all') params.set('status', selectedStatus);
		if (typeof window !== 'undefined') window.history.replaceState({}, '', `${window.location.pathname}${params.toString() ? `?${params}` : ''}`);
	}
	$: attentionGroups = [...new Map(attention.map((item) => [item.itemID, { ...item, signals: attention.filter((signal) => signal.itemID === item.itemID) }])).values()];
	$: baseFilteredItems = selectedFilter === 'active'
		? activeItems
		: selectedFilter === 'planned'
			? plannedItems
			: selectedFilter === 'capabilities'
				? capabilityItems
				: selectedFilter === 'history'
					? historyItems
			: roadmap.items;
	$: filteredItems = baseFilteredItems.filter((item) => (selectedHorizon === 'all' || item.horizon === selectedHorizon) && (selectedStatus === 'all' || item.status === selectedStatus) && (!search || `${item.id} ${item.title} ${item.description}`.toLowerCase().includes(search.toLowerCase())));
</script>

<svelte:head>
	<title>{roadmap.project_name} · local dashboard</title>
	<meta name="description" content="A compact local view of active Cassor work, decisions, capabilities, and delivery history." />
</svelte:head>

<main class="mx-auto max-w-6xl px-5 py-8 sm:px-8 lg:py-12">
	<header class="flex flex-col gap-5 border-b border-slate-800 pb-8 lg:flex-row lg:items-end lg:justify-between">
		<div>
			<p class="text-xs font-semibold tracking-[0.22em] text-cassor-400">CASSOR · LOCAL WORKBOARD</p>
			<h1 class="mt-2 text-3xl font-bold tracking-tight text-white sm:text-5xl">{roadmap.project_name}</h1>
			<p class="mt-3 max-w-2xl text-base leading-7 text-slate-300">See what needs attention, what is next, and what has been delivered.</p>
		</div>
		<div class="text-sm text-slate-400 lg:text-right">
			<p>Generated {generatedLabel}</p>
			<p class="mt-1 text-emerald-300">{shippedFeatureCount} delivered items</p>
		</div>
	</header>

	<section class="mt-8" aria-labelledby="attention-title">
		<div class="flex items-center justify-between gap-4"><div><h2 id="attention-title" class="text-xl font-bold text-white">Needs attention</h2><p class="mt-1 text-sm text-slate-400">Signals grouped by roadmap item.</p></div><span class="rounded-full border border-amber-400/30 bg-amber-400/10 px-3 py-1 text-sm font-semibold text-amber-200">{attentionGroups.length}</span></div>
		{#if attention.length > 0}
			<div class="mt-4 grid gap-3 md:grid-cols-2">{#each attentionGroups as item}<a class={`rounded-xl border p-4 transition hover:-translate-y-0.5 hover:border-cyan-300 ${item.tone}`} href={item.href}><p class="text-xs font-semibold uppercase tracking-wider">RM-{item.itemID} · {item.signals.length} signal{item.signals.length === 1 ? '' : 's'}</p><p class="mt-2 text-sm font-semibold text-white">{item.label}</p><p class="mt-1 text-sm text-slate-200">{item.detail}</p>{#if item.signals.length > 1}<p class="mt-2 text-xs text-slate-400">{item.signals.slice(1).map((signal) => signal.detail).join(' · ')}</p>{/if}</a>{/each}</div>
		{:else}<p class="mt-4 rounded-xl border border-emerald-400/30 bg-emerald-400/10 p-4 text-sm text-emerald-200">Nothing needs attention right now.</p>{/if}
	</section>

	<section class="mt-10" aria-labelledby="work-title">
		<div class="flex flex-col gap-4"><div class="flex flex-wrap gap-2" role="tablist" aria-label="Roadmap views">{#each [['active', 'Active'], ['planned', 'Planned'], ['capabilities', 'Capabilities'], ['history', 'History'], ['all', 'All']] as [value, label]}<button type="button" role="tab" aria-selected={selectedFilter === value} class={`rounded-full px-4 py-2 text-sm font-semibold transition ${selectedFilter === value ? 'bg-cyan-400 text-slate-950' : 'border border-slate-700 text-slate-300 hover:border-cyan-400 hover:text-cyan-200'}`} onclick={() => { selectedFilter = value as Filter; updateURL(); }}>{label}</button>{/each}</div><div class="grid gap-3 sm:grid-cols-[minmax(12rem,1fr)_auto_auto]"><label class="sr-only" for="roadmap-search">Filter roadmap items</label><input id="roadmap-search" bind:value={search} oninput={updateURL} placeholder="Filter by title or ID" class="rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-slate-200 placeholder:text-slate-500" /><select aria-label="Filter by horizon" bind:value={selectedHorizon} onchange={updateURL} class="rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-slate-200"><option value="all">All horizons</option><option value="Now">Now</option><option value="Next">Next</option><option value="Later">Later</option></select><select aria-label="Filter by status" bind:value={selectedStatus} onchange={updateURL} class="rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-slate-200"><option value="all">All statuses</option>{#each [...new Set(roadmap.items.map((item) => item.status))] as status}<option value={status}>{status}</option>{/each}</select></div></div>
		<div class="mt-4 divide-y divide-slate-800 overflow-hidden rounded-xl border border-slate-800 bg-slate-900/60">{#if filteredItems.length > 0}{#each filteredItems as item}<a class="flex flex-col gap-2 px-4 py-4 transition hover:bg-slate-800/70 sm:flex-row sm:items-center sm:justify-between" href={`${base}/roadmap/rm-${item.id}/`}><div><p class="text-xs font-semibold tracking-wider text-cyan-300">RM-{item.id} · {item.feature_type}</p><h3 class="mt-1 font-semibold text-white">{item.title}</h3><p class="mt-1 line-clamp-2 text-sm text-slate-400">{item.description || item.current_state || item.rationale || 'No summary recorded.'}</p></div><div class="flex items-center gap-3 text-sm"><span class="text-slate-400">{item.horizon}</span><span class="rounded-full border border-slate-700 px-3 py-1 text-slate-200">{item.status}</span></div></a>{/each}{:else}<p class="p-5 text-sm text-slate-400">No items match this view.</p>{/if}</div>
	</section>

	<section class="mt-10" aria-labelledby="map-title">
		<div class="flex items-center justify-between gap-4"><div><h2 id="map-title" class="text-xl font-bold text-white">Product map</h2><p class="mt-1 text-sm text-slate-400">Capabilities explain what Cassor can do; gaps explain what it cannot yet do.</p></div><a class="text-sm font-semibold text-cyan-300 hover:text-cyan-200" href={`${base}/application/`}>View application map</a></div>
		<div class="mt-4 grid gap-3 sm:grid-cols-3"><div class="rounded-xl border border-slate-800 bg-slate-900/60 p-4"><p class="text-sm text-slate-400">Capabilities</p><p class="mt-2 text-2xl font-bold text-emerald-300">{capabilityItems.length}</p></div><div class="rounded-xl border border-slate-800 bg-slate-900/60 p-4"><p class="text-sm text-slate-400">Active work</p><p class="mt-2 text-2xl font-bold text-cyan-300">{activeItems.length}</p></div><div class="rounded-xl border border-slate-800 bg-slate-900/60 p-4"><p class="text-sm text-slate-400">Delivered</p><p class="mt-2 text-2xl font-bold text-white">{historyItems.length}</p></div></div>
	</section>

	<section class="mt-12" aria-labelledby="history-title">
		<div class="flex items-center justify-between gap-4"><div><h2 id="history-title" class="text-xl font-bold text-white">Delivery history</h2><p class="mt-1 text-sm text-slate-400">The timeline remains available for reviewing what changed.</p></div><span class="text-sm text-slate-400">{scheduled.length} dated milestones</span></div>
		<div class="relative mt-6 border-l border-slate-700 pl-6">{#each scheduled as item}<a class="group relative mb-3 block rounded-xl border border-slate-800 bg-slate-900/50 p-4 transition hover:border-cyan-400 hover:bg-slate-900" href={`${base}/roadmap/rm-${item.id}/`}><span class={`absolute -left-[1.92rem] top-5 h-3 w-3 rounded-full border-2 border-cassor-950 ${milestoneKind(item) === 'done' ? 'bg-emerald-400' : 'bg-cyan-400'}`}></span><div class="flex flex-wrap items-center gap-x-3 gap-y-1"><span class="text-xs font-semibold tracking-wider text-cyan-300">{milestoneDate(item)} · RM-{item.id}</span><span class="text-xs text-slate-500">{item.status}</span></div><h3 class="mt-1 font-semibold text-white group-hover:text-cyan-200">{item.title}</h3></a>{/each}</div>
		{#if unscheduled.length > 0}<h3 class="mt-8 text-sm font-semibold uppercase tracking-wider text-slate-400">Unscheduled</h3><div class="mt-3 grid gap-3 sm:grid-cols-2">{#each unscheduled as item}<a class="rounded-xl border border-slate-800 bg-slate-900/50 p-4 transition hover:border-cyan-400" href={`${base}/roadmap/rm-${item.id}/`}><p class="text-xs text-slate-500">RM-{item.id} · {item.status}</p><p class="mt-1 font-semibold text-white">{item.title}</p></a>{/each}</div>{/if}
	</section>
</main>
