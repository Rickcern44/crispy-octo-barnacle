import adapter from "@sveltejs/adapter-static";
import { mdsvex } from "mdsvex";

/** @type {import('@sveltejs/kit').Config} */
const config = {
	extensions: [".svelte", ".svx", ".md"],
	preprocess: [mdsvex({ extensions: [".svx", ".md"] })],
	kit: {
		adapter: adapter({
			pages: "../docs/roadmap",
			assets: "../docs/roadmap",
			fallback: "404.html",
		}),
		paths: { base: process.env.SITE_BASE ?? "" },
	},
};

export default config;
