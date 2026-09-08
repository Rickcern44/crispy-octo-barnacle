<script lang="ts">
	import { base } from '$app/paths';
	import { page } from '$app/state';
	import { artifactsForItem, capabilityForChange, changesForCapability, displayProgress, lifecycleForItem, plansForItem, relationshipsForItem, reportsForItem, roadmap, specificationsForItem, stateHistoryForCapability } from '$lib/roadmap';

	const itemID = Number((page.params.id ?? '').replace('rm-', ''));
	const item = roadmap.items.find((candidate) => candidate.id === itemID);
	const plans = plansForItem(itemID);
	const reports = reportsForItem(itemID);
	const lifecycle = lifecycleForItem(itemID);
	const relationships = relationshipsForItem(itemID);
	const specifications = item ? specificationsForItem(item) : [];
	const linkedChanges = item?.feature_type === 'Capability' ? changesForCapability(itemID) : [];
	const linkedCapabilities = item?.feature_type === 'Change' ? capabilityForChange(itemID) : [];
	const stateHistory = item?.feature_type === 'Capability' ? stateHistoryForCapability(itemID) : [];
	const artifacts = artifactsForItem(itemID);
</script>

<svelte:head><title>{item ? `RM-${item.id} · ${item.title}` : 'Roadmap item'}</title></svelte:head>

<main class="mx-auto max-w-4xl px-5 py-12 sm:px-8 lg:py-20">
	<a class="text-sm font-semibold text-cyan-300 hover:text-cyan-200" href={`${base}/`}>← Back to roadmap</a>
	{#if item}
		<header class="mt-8 border-b border-slate-800 pb-8"><p class="text-sm font-semibold tracking-[0.18em] text-cyan-400">RM-{item.id} · {item.feature_type} · {item.status}</p><h1 class="mt-3 text-4xl font-bold tracking-tight text-white sm:text-5xl">{item.title}</h1><p class="mt-5 text-lg leading-8 text-slate-300">{item.technical_summary || item.description || item.rationale}</p></header>
		<section class="mt-9 grid gap-4 sm:grid-cols-2"><div class="rounded-xl border border-slate-800 bg-slate-900/70 p-5"><p class="text-xs font-semibold uppercase tracking-wider text-slate-400">Delivery</p><dl class="mt-3 space-y-2 text-sm"><div class="flex justify-between gap-3"><dt class="text-slate-400">Progress</dt><dd>{displayProgress(item)}%</dd></div><div class="flex justify-between gap-3"><dt class="text-slate-400">Target</dt><dd>{item.target_date || 'Unscheduled'}</dd></div><div class="flex justify-between gap-3"><dt class="text-slate-400">Horizon</dt><dd>{item.horizon}</dd></div></dl></div><div class="rounded-xl border border-slate-800 bg-slate-900/70 p-5"><p class="text-xs font-semibold uppercase tracking-wider text-slate-400">Ownership</p><dl class="mt-3 space-y-2 text-sm"><div class="flex justify-between gap-3"><dt class="text-slate-400">Team</dt><dd>{item.team || 'Unassigned'}</dd></div><div class="flex justify-between gap-3"><dt class="text-slate-400">Lead</dt><dd>{item.lead_engineer || 'Unassigned'}</dd></div><div class="flex justify-between gap-3"><dt class="text-slate-400">Priority</dt><dd>{item.priority || 'Unassigned'}</dd></div></dl></div></section>
		<section class="mt-9 rounded-xl border border-slate-800 bg-slate-900/70 p-5"><p class="text-xs font-semibold uppercase tracking-wider text-slate-400">Current product state</p><p class="mt-3 leading-7 text-slate-300">{item.current_state || 'Not yet documented.'}</p></section>
		{#if linkedChanges.length > 0}<section class="mt-9"><h2 class="text-2xl font-bold text-white">Delivery changes for this capability</h2><ul class="mt-4 space-y-3">{#each linkedChanges as link}<li class="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><a class="font-medium text-cyan-300 hover:text-cyan-200" href={`${base}/roadmap/rm-${link.change_item_id}/`}>{link.change_item_title}</a><span class="ml-2 text-sm text-slate-400">linked delivery change</span></li>{/each}</ul></section>{/if}
		{#if linkedCapabilities.length > 0}<section class="mt-9"><h2 class="text-2xl font-bold text-white">Capability dossier</h2><ul class="mt-4 space-y-3">{#each linkedCapabilities as link}<li class="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><span class="text-sm text-slate-400">Change of </span><a class="font-medium text-cyan-300 hover:text-cyan-200" href={`${base}/roadmap/rm-${link.capability_item_id}/`}>{link.capability_title}</a></li>{/each}</ul></section>{/if}
		{#if stateHistory.length > 0}<section class="mt-9"><h2 class="text-2xl font-bold text-white">Accepted state history</h2><ul class="mt-4 space-y-3">{#each stateHistory as state}<li class="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><p class="text-sm text-slate-300">{state.state}</p><p class="mt-2 text-xs text-slate-500">Accepted by {state.accepted_by} · {state.accepted_at.slice(0, 10)}{#if state.source_change_item_id} · source change RM-{state.source_change_item_id}{/if}</p></li>{/each}</ul></section>{/if}
		{#if artifacts.length > 0}<section class="mt-9"><h2 class="text-2xl font-bold text-white">Dossier findings</h2><ul class="mt-4 space-y-3">{#each artifacts as artifact}<li class="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><div class="flex flex-wrap items-center gap-2"><span class="font-medium text-white">{artifact.kind}</span><span class="text-xs uppercase tracking-wider text-cyan-300">{artifact.status}</span></div><p class="mt-2 text-sm text-slate-300">{artifact.summary}</p><p class="mt-2 text-xs text-slate-500">{artifact.author_role}</p></li>{/each}</ul></section>{/if}
		{#if relationships.length > 0}<section class="mt-9"><h2 class="text-2xl font-bold text-white">Feature relationships</h2><ul class="mt-4 space-y-3">{#each relationships as relationship}<li class="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><a class="font-medium text-cyan-300 hover:text-cyan-200" href={`${base}/roadmap/rm-${relationship.source_item_id === itemID ? relationship.target_item_id : relationship.source_item_id}/`}>{relationship.source_item_id === itemID ? `${relationship.relationship_type} → ${relationship.target_item_title}` : `${relationship.source_item_title} → ${relationship.relationship_type}`}</a></li>{/each}</ul></section>{/if}
		{#if specifications.length > 0}<section class="mt-9"><h2 class="text-2xl font-bold text-white">Technical specifications</h2><ul class="mt-4 space-y-2">{#each specifications as specification}<li class="rounded-lg border border-slate-800 bg-slate-900/70 p-3 text-sm text-slate-300"><span class={specification.done ? 'text-emerald-300' : 'text-cyan-300'}>{specification.done ? '✓' : '○'}</span> <span class="font-medium text-white">{specification.title}</span>{#if specification.description}<p class="mt-1 text-slate-400">{specification.description}</p>{/if}</li>{/each}</ul></section>{/if}
		{#if plans.length > 0}
			<section class="mt-9">
				<h2 class="text-2xl font-bold text-white">Implementation history</h2>
				<p class="mt-2 text-slate-400">Approved plan tasks and their recorded outcomes.</p>
				<div class="mt-5 space-y-5">
					{#each plans as plan}
						<div class="rounded-xl border border-slate-800 bg-slate-900/70 p-5">
							<p class="text-xs font-semibold uppercase tracking-wider text-cyan-300">Plan revision {plan.revision} · {plan.status}</p>
							<ul class="mt-4 space-y-4">
								{#each plan.tasks as task}
									<li class="border-l-2 border-slate-700 pl-4">
										<div class="flex flex-wrap items-center gap-x-3 gap-y-1"><h3 class="font-semibold text-white">{task.title}</h3><span class="text-xs font-medium uppercase tracking-wider text-slate-400">{task.status}</span>{#if task.completed_at}<span class="text-xs text-emerald-300">Completed {task.completed_at.slice(0, 10)}</span>{/if}</div>
										{#if task.description}<p class="mt-1 text-sm text-slate-300">{task.description}</p>{/if}
										{#if task.outcome}<p class="mt-2 text-sm text-slate-400">{task.outcome}</p>{/if}
									</li>
								{/each}
							</ul>
						</div>
					{/each}
				</div>
			</section>
		{/if}
		{#if reports.length > 0}
			<section class="mt-9"><h2 class="text-2xl font-bold text-white">Orchestration reports</h2><div class="mt-5 space-y-4">{#each reports as report}<article class="rounded-xl border border-slate-800 bg-slate-900/70 p-5"><div class="flex flex-wrap justify-between gap-2"><h3 class="font-semibold text-white">{report.execution_mode} run</h3><span class="text-sm text-slate-400">{new Date(report.created_at).toLocaleDateString()}</span></div><p class="mt-2 text-sm text-slate-300">Roles: {report.roles.join(', ') || 'none'} · {Math.round(report.elapsed_ns / 1_000_000_000)}s · {report.tool_calls} tool calls · {report.verification}</p><p class="mt-2 text-sm text-slate-400">Tokens: {report.total_tokens ?? 'unavailable'}</p></article>{/each}</div></section>
		{/if}
		<section class="mt-9"><h2 class="text-2xl font-bold text-white">SDD-lite lifecycle</h2><div class="mt-4 flex flex-wrap gap-2">{#each ['Intake', 'Explore', 'Define', 'Plan', 'Implement', 'Verify', 'Record'] as phase}<span class={`rounded-full px-3 py-1 text-sm ${lifecycle.phases.some((record) => record.phase === phase) ? 'bg-cyan-400 text-slate-950' : 'bg-slate-800 text-slate-400'}`}>{phase}</span>{/each}</div>{#if lifecycle.criteria.length > 0}<h3 class="mt-6 font-semibold text-white">Acceptance criteria</h3><ul class="mt-3 space-y-2">{#each lifecycle.criteria as criterion}<li class="rounded-lg border border-slate-800 bg-slate-900/70 p-3"><span class="font-medium text-white">{criterion.Title}</span><span class="ml-2 text-sm text-cyan-300">{criterion.Status}</span>{#if criterion.Evidence}<p class="mt-1 text-sm text-slate-400">{criterion.Evidence}</p>{/if}</li>{/each}</ul>{/if}</section>
	{:else}<p class="mt-8 text-slate-300">This roadmap item does not exist.</p>{/if}
</main>
