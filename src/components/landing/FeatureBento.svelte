<script lang="ts">
	import type { CorpusStats } from '$server/db';

	let { stats }: { stats: CorpusStats } = $props();

	const n = (v: number) => v.toLocaleString('en-US');
</script>

<div class="grid gap-4 md:grid-cols-6">
	<article class="panel-card flex flex-col justify-between gap-6 p-6 md:col-span-4 md:rounded-3xl">
		<div>
			<h3 class="text-body-lg font-medium">Six export formats, generated not written</h3>
			<p class="mt-2 max-w-lg text-body text-muted-foreground">
				DESIGN.md, CSS custom properties, a Tailwind v4 theme, W3C design tokens, the structured
				system as JSON, and the prompt snapshot. Compact and extended where it matters.
			</p>
		</div>

		<svg viewBox="0 0 320 84" class="h-20 w-full" aria-hidden="true">
			{#each [0, 1, 2, 3, 4] as i (i)}
				<rect
					x={i * 64}
					y={10 + i * 2}
					width="54"
					height="64"
					rx="8"
					class="fill-muted stroke-border"
					stroke-width="1"
				/>
				<rect x={i * 64 + 10} y={24 + i * 2} width="26" height="4" rx="2" class="fill-primary" />
				<rect x={i * 64 + 10} y={34 + i * 2} width="34" height="3" rx="1.5" class="fill-border-strong" />
				<rect x={i * 64 + 10} y={42 + i * 2} width="20" height="3" rx="1.5" class="fill-border-strong" />
			{/each}
		</svg>
	</article>

	<article class="panel-card flex flex-col justify-between gap-6 p-6 md:col-span-2 md:rounded-3xl">
		<div>
			<h3 class="text-body-lg font-medium">Measured, not guessed</h3>
			<p class="mt-2 text-body text-muted-foreground">
				Every colour carries frequency, prominence, OKLCH and the contexts it appears in.
			</p>
		</div>
		<svg viewBox="0 0 160 64" class="h-16 w-full" aria-hidden="true">
			{#each [40, 26, 54, 18, 44, 32] as h, i (i)}
				<rect x={i * 26} y={64 - h} width="16" height={h} rx="4" class="fill-primary" opacity={0.35 + i * 0.1} />
			{/each}
		</svg>
	</article>

	<article class="panel-card flex flex-col justify-between gap-6 p-6 md:col-span-2 md:rounded-3xl">
		<div>
			<h3 class="text-body-lg font-medium">{n(stats.colors)} colours, deduplicated</h3>
			<p class="mt-2 text-body text-muted-foreground">
				One row per distinct value across the corpus, so you can ask which systems share a colour.
			</p>
		</div>
		<svg viewBox="0 0 160 48" class="h-12 w-full" aria-hidden="true">
			{#each Array(16) as _, i (i)}
				<rect
					x={(i % 8) * 20}
					y={i < 8 ? 0 : 26}
					width="16"
					height="18"
					rx="4"
					class={i % 3 === 0 ? 'fill-primary' : 'fill-muted stroke-border'}
					stroke-width="1"
				/>
			{/each}
		</svg>
	</article>

	<article class="panel-card flex flex-col justify-between gap-6 p-6 md:col-span-4 md:rounded-3xl">
		<div>
			<h3 class="text-body-lg font-medium">
				{n(stats.fonts)} font families with the fallback each one names
			</h3>
			<p class="mt-2 max-w-lg text-body text-muted-foreground">
				Weights, size ramps, line heights, letter spacing and OpenType features, plus the substitute
				to reach for when the original is not licensed to you.
			</p>
		</div>
		<svg viewBox="0 0 320 60" class="h-14 w-full" aria-hidden="true">
			{#each [{ w: 240, y: 6, h: 14 }, { w: 180, y: 26, h: 10 }, { w: 120, y: 42, h: 7 }] as bar (bar.y)}
				<rect x="0" y={bar.y} width={bar.w} height={bar.h} rx={bar.h / 2} class="fill-border-strong" />
			{/each}
			<rect x="252" y="6" width="56" height="14" rx="7" class="fill-primary" />
		</svg>
	</article>
</div>
