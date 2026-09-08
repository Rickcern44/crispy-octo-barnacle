import { readFile } from "node:fs/promises";
import { resolve } from "node:path";

const root = resolve(import.meta.dirname, "..");
const files = {
	roadmap: await readFile(resolve(root, "src/lib/roadmap.ts"), "utf8"),
	timeline: await readFile(resolve(root, "src/routes/+page.svelte"), "utf8"),
	detail: await readFile(resolve(root, "src/routes/roadmap/[id]/+page.svelte"), "utf8"),
};

const expectations = [
	[files.roadmap, 'return item.status === "Done" ? 100 : item.progress;', "shared Done-item progress rule"],
	[files.detail, "{displayProgress(item)}%", "detail-page progress display"],
	[files.roadmap, "export function specificationsForItem", "specification parsing helper"],
	[files.roadmap, "export const shippedFeatureCount", "derived shipped-feature count"],
	[files.timeline, "{shippedFeatureCount} features shipped", "dashboard shipped counter"],
	[files.detail, "{#if specifications.length > 0}", "conditional specification section"],
];

for (const [source, expected, label] of expectations) {
	if (!source.includes(expected)) {
		throw new Error(`Missing ${label}: ${expected}`);
	}
}

if (files.detail.includes("{item.specifications}")) {
	throw new Error("Technical specifications must not render raw JSON");
}

console.log("Roadmap dashboard presentation contract passed");
