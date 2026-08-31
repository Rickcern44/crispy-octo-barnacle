<script lang="ts">
	import { onMount } from "svelte";
	import { base } from "$app/paths";

	let element: HTMLDivElement;

	onMount(async () => {
		const stylesheet = document.createElement("link");
		stylesheet.rel = "stylesheet";
		stylesheet.href = `${base}/pagefind/pagefind-ui.css`;
		document.head.append(stylesheet);
		const pagefind = await import(/* @vite-ignore */ `${base}/pagefind/pagefind-ui.js`);
		new pagefind.PagefindUI({ element, showSubResults: true });
	});
</script>

<div bind:this={element} class="cassor-search"></div>
