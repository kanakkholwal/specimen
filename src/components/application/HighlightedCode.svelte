<script lang="ts">
	import type { HighlighterCore } from 'shiki/core';
	import { cn } from '$lib/utils';

	type Props = { code: string; lang: string; class?: string };
	let { code, lang, class: className }: Props = $props();

	let html = $state('');

	// Shiki is ~1MB of grammars, so it loads on demand and only for the three languages
	// these exports actually use. Plain text renders until it arrives.
	let loader: Promise<HighlighterCore> | null = null;

	function highlighter() {
		loader ??= (async () => {
			const [core, engine, md, css, json, themes] = await Promise.all([
				import('shiki/core'),
				import('shiki/engine/javascript'),
				import('shiki/langs/markdown.mjs'),
				import('shiki/langs/css.mjs'),
				import('shiki/langs/json.mjs'),
				import('$lib/shiki-jetbrains')
			]);
			return core.createHighlighterCore({
				engine: engine.createJavaScriptRegexEngine(),
				langs: [md.default, css.default, json.default],
				themes: [themes.intellijLight, themes.darcula]
			});
		})();
		return loader;
	}

	$effect(() => {
		const source = code;
		const language = lang;
		let cancelled = false;

		highlighter()
			.then((hl) => {
				if (cancelled) return;
				html = hl.codeToHtml(source, {
					lang: language,
					themes: { light: 'jetbrains-light', dark: 'jetbrains-darcula' },
					defaultColor: false
				});
			})
			.catch(() => {
				if (!cancelled) html = '';
			});

		return () => {
			cancelled = true;
		};
	});
</script>

{#if html}
	<div class={cn('shiki-host', className)}>
		<!-- eslint-disable-next-line svelte/no-at-html-tags -->
		{@html html}
	</div>
{:else}
	<pre class={cn('shiki-host', className)}><code>{code}</code></pre>
{/if}

<style>
	.shiki-host :global(pre) {
		margin: 0;
		background-color: transparent !important;
		padding: 0;
	}

	.shiki-host :global(code) {
		display: block;
		font-family: var(--font-code);
	}

	/* defaultColor false emits both themes; this picks the one in use. */
	.shiki-host :global(.shiki span) {
		color: var(--shiki-light);
	}

	:global(.dark) .shiki-host :global(.shiki span) {
		color: var(--shiki-dark);
	}
</style>
