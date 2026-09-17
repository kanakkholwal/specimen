import adapter from '@sveltejs/adapter-cloudflare';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter(),
		// The dev server watches and locks its own output, so the gate builds into a sibling.
		outDir: process.env.SVELTE_KIT_OUT_DIR || '.svelte-kit',
		alias: {
			$components: 'src/components',
			$lib: 'src/lib',
			$server: 'src/lib/server'
		}
	}
};

export default config;
