import { roadmap } from "$lib/roadmap";

export const entries = () =>
	roadmap.items.map((item) => ({ id: `rm-${item.id}` }));
export const prerender = true;
