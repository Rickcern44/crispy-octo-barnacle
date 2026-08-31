<script lang="ts">
	import { base } from '$app/paths';
	import favicon from '$lib/assets/favicon.svg';
	import Search from '$lib/Search.svelte';
	import './layout.css';

	let { children } = $props();
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>
<header class="border-b border-slate-800 bg-cassor-950/95">
	<div class="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-4 px-5 py-4 sm:px-8">
		<a class="font-bold tracking-tight text-white" href={`${base}/`}>Cassor</a>
		<nav class="flex items-center gap-4 text-sm font-medium text-slate-300" aria-label="Primary navigation"><a class="hover:text-cyan-300" href={`${base}/`}>Roadmap</a><a class="hover:text-cyan-300" href={`${base}/guides/project-handoff/`}>Docs</a></nav>
		<div class="w-full sm:w-72"><Search /></div>
	</div>
</header>
{@render children()}
