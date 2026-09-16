<script lang="ts">
	import { IconCheck, IconCopy } from '@tabler/icons-svelte';
	import { type ColorFormat, contrast, formatColor, slug } from '$lib/color';
	import type { NamedColor } from '$server/db';
	import { cn } from '$lib/utils';

	let { colors }: { colors: NamedColor[] } = $props();

	const FORMATS: { id: ColorFormat; label: string }[] = [
		{ id: 'hex', label: 'HEX' },
		{ id: 'rgb', label: 'RGB' },
		{ id: 'oklch', label: 'OKLCH' },
		{ id: 'var', label: 'CSS var' }
	];

	let format = $state<ColorFormat>('hex');
	let copiedKey = $state('');
	let announce = $state('');

	const groups = $derived(
		[...new Set(colors.map((c) => c.group ?? 'other'))].map((group) => ({
			group,
			items: colors.filter((c) => (c.group ?? 'other') === group)
		}))
	);

	async function copy(value: string, key: string) {
		try {
			await navigator.clipboard.writeText(value);
			copiedKey = key;
			announce = `Copied ${value}`;
			setTimeout(() => {
				if (copiedKey === key) copiedKey = '';
			}, 1600);
		} catch {
			announce = 'Copying was blocked by the browser.';
		}
	}

	function copyAll() {
		const sheet = colors
			.map((c) => `  --color-${slug(c.name ?? c.hex)}: ${formatColor(format, c)};`)
			.join('\n');
		copy(`:root {\n${sheet}\n}`, 'all');
	}

	// A swatch can be any colour, including the page background, so the chip carries a
	// hairline and the readable ratio is stated rather than implied.
	function onLight(hex: string) {
		return contrast(hex, '#ffffff');
	}
	function onDark(hex: string) {
		return contrast(hex, '#0a0a0a');
	}
</script>

<div class="flex flex-wrap items-center justify-between gap-3">
	<div
		role="group"
		aria-label="Colour format"
		class="flex items-center rounded-md border border-border p-0.5"
	>
		{#each FORMATS as f (f.id)}
			<button
				type="button"
				onclick={() => (format = f.id)}
				aria-pressed={format === f.id}
				class={cn(
					'h-8 rounded-sm px-2.5 text-caption transition-colors',
					'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring',
					format === f.id
						? 'bg-muted font-medium text-foreground'
						: 'text-muted-foreground hover:text-foreground'
				)}
			>
				{f.label}
			</button>
		{/each}
	</div>

	<button
		type="button"
		onclick={copyAll}
		class="inline-flex h-10 items-center gap-2 rounded-md border border-border px-3 text-body text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
	>
		{#if copiedKey === 'all'}
			<IconCheck size={15} stroke={2} aria-hidden="true" />
			Copied palette
		{:else}
			<IconCopy size={15} stroke={1.75} aria-hidden="true" />
			Copy all as CSS
		{/if}
	</button>
</div>

<p aria-live="polite" class="sr-only">{announce}</p>

<div class="mt-8 flex flex-col gap-10">
	{#each groups as { group, items } (group)}
		<section>
			<h3 class="label-eyebrow uppercase text-muted-foreground">{group}</h3>

			<ul class="mt-4 grid grid-cols-2 gap-x-5 gap-y-8 md:grid-cols-3 xl:grid-cols-4">
				{#each items as color (color.hex + color.name)}
					{@const key = color.hex + color.name}
					{@const value = formatColor(format, color)}
					<li>
						<button
							type="button"
							onclick={() => copy(value, key)}
							class="group block w-full text-left focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-ring"
						>
							<span
								class="relative flex h-24 items-start justify-end rounded-xl border border-border-strong p-2"
								style:background-color={color.hex}
							>
								<span
									class={cn(
										'inline-flex items-center gap-1.5 rounded-md bg-card px-2 py-1 text-caption text-foreground shadow-xs transition-opacity',
										copiedKey === key
											? 'opacity-100'
											: 'opacity-0 group-hover:opacity-100 group-focus-visible:opacity-100'
									)}
								>
									{#if copiedKey === key}
										<IconCheck size={13} stroke={2.5} aria-hidden="true" />
										Copied
									{:else}
										<IconCopy size={13} stroke={1.75} aria-hidden="true" />
										Copy
									{/if}
								</span>
							</span>

							<span class="mt-3 block text-body font-medium">{color.name ?? color.hex}</span>
							<span class="mt-0.5 block font-mono text-caption text-muted-foreground">{value}</span>

							<span class="mt-2 flex flex-wrap gap-1.5">
								{#each [{ label: 'on light', ratio: onLight(color.hex) }, { label: 'on dark', ratio: onDark(color.hex) }] as r (r.label)}
									<span
										class="inline-flex items-center gap-1 rounded-xs border border-border px-1.5 py-0.5 text-caption text-muted-foreground"
									>
										<span class="font-mono">{r.ratio.toFixed(1)}</span>
										<span>{r.label}</span>
										<span class="font-medium text-foreground">
											{r.ratio >= 4.5 ? 'AA' : r.ratio >= 3 ? 'UI' : 'low'}
										</span>
									</span>
								{/each}
							</span>

							{#if color.role}
								<span class="mt-2 block text-caption leading-relaxed text-muted-foreground">
									{color.role}
								</span>
							{/if}
						</button>
					</li>
				{/each}
			</ul>
		</section>
	{/each}
</div>
