<script lang="ts">
	import { onMount } from "svelte";
	import { base } from "$app/paths";

	let element: HTMLDivElement;

	onMount(async () => {
		const stylesheet = document.createElement("link");
		stylesheet.rel = "stylesheet";
		stylesheet.href = `${base}/pagefind/pagefind-ui.css`;
		document.head.append(stylesheet);

		const script = document.createElement("script");
		script.src = `${base}/pagefind/pagefind-ui.js`;
		script.async = true;
		document.body.append(script);
		await new Promise<void>((resolve, reject) => {
			script.addEventListener("load", () => resolve(), { once: true });
			script.addEventListener("error", () => reject(new Error("Unable to load Pagefind search")), { once: true });
		});
		const PagefindUI = (window as typeof window & { PagefindUI?: new (options: { element: HTMLDivElement; showSubResults: boolean }) => unknown }).PagefindUI;
		if (!PagefindUI) throw new Error("Pagefind search UI was not exposed");
		new PagefindUI({ element, showSubResults: true });
	});
</script>

<div bind:this={element} class="cassor-search"></div>
