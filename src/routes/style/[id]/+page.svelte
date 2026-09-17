<script lang="ts">
	import {
		IconArrowUpRight,
		IconCheck,
		IconExternalLink,
		IconMoon,
		IconSun,
		IconX
	} from '@tabler/icons-svelte';
	import ColorPalette from '$components/application/ColorPalette.svelte';
	import ComponentGallery from '$components/application/ComponentGallery.svelte';
	import ExportPanel from '$components/application/ExportPanel.svelte';
	import TypographySpecimen from '$components/application/TypographySpecimen.svelte';
	import RailFrame from '$components/site/RailFrame.svelte';
	import RailRow from '$components/site/RailRow.svelte';
	import { Button } from '$components/ui/button';
	import { site } from '$lib/constants';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const s = $derived(data.style);
	const dos = $derived(data.guidelines.filter((g) => g.kind === 'do'));
	const donts = $derived(data.guidelines.filter((g) => g.kind === 'dont'));
	const screenshot = $derived(data.media.find((m) => m.role === 'screenshot')?.url ?? null);
	const theme = $derived(s.colorScheme ?? s.theme ?? '');

</script>

<svelte:head>
	<title>{s.siteName} design system | {site.name}</title>
	<meta
		name="description"
		content={s.description ?? `The extracted design system for ${s.siteName}.`}
	/>
</svelte:head>

<RailFrame>
	<RailRow divider={false} label="Overview" class="pt-28 pb-8 sm:pt-32">
		<div class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(0,520px)] lg:gap-12">
			<div>
				<p class="text-caption text-muted-foreground">
					<a href="/sites/{s.origin}" class="hover:text-foreground">{s.origin}</a>
				</p>
				<h1 class="mt-3 text-heading-lg font-medium md:text-display">{s.siteName}</h1>
				{#if s.northStar}
					<p class="mt-4 text-body-lg text-primary">{s.northStar}</p>
				{/if}
				{#if s.description}
					<p class="mt-4 max-w-2xl text-body leading-relaxed text-muted-foreground">
						{s.description}
					</p>
				{/if}

				<ul class="mt-6 flex flex-wrap gap-2">
					{#if theme}
						<li
							class="inline-flex items-center gap-1.5 rounded-full border border-border px-3 py-1 text-caption text-muted-foreground"
						>
							{#if theme === 'dark'}
								<IconMoon size={12} stroke={1.75} aria-hidden="true" />
							{:else}
								<IconSun size={12} stroke={1.75} aria-hidden="true" />
							{/if}
							{theme}
						</li>
					{/if}
					{#if s.industry}
						<li
							class="rounded-full border border-border px-3 py-1 text-caption text-muted-foreground"
						>
							{s.industry}
						</li>
					{/if}
					{#if s.density}
						<li
							class="rounded-full border border-border px-3 py-1 text-caption text-muted-foreground"
						>
							{s.density} spacing
						</li>
					{/if}
					{#if s.scaleName}
						<li
							class="rounded-full border border-border px-3 py-1 text-caption text-muted-foreground"
						>
							{s.scaleName} scale
						</li>
					{/if}
				</ul>

				<div class="mt-8 flex flex-wrap gap-3">
					<Button href={s.url} target="_blank" rel="noreferrer noopener" variant="default">
						Visit {s.origin}
						<IconExternalLink size={16} stroke={1.75} aria-hidden="true" />
					</Button>
				</div>
			</div>

			{#if screenshot}
				<figure class="panel-card overflow-hidden">
					<img
						src={screenshot}
						alt="Screenshot of {s.siteName}"
						loading="lazy"
						decoding="async"
						class="w-full object-cover object-top"
					/>
				</figure>
			{/if}
		</div>
	</RailRow>

	{#if data.artifactNames.length}
		<RailRow label="Exports" class="py-8 sm:py-10">
			<div class="flex flex-wrap items-end justify-between gap-3">
				<div>
					<h2 class="text-heading-sm font-medium">Copy this system</h2>
					<p class="mt-2 text-body text-muted-foreground">
						Generated from the measured extraction, byte for byte identical to the source tool.
					</p>
				</div>
			</div>
			<div class="mt-6">
				<ExportPanel
					styleId={data.style.id}
					available={data.artifactNames}
					initial={data.initialExport}
				/>
			</div>
		</RailRow>
	{/if}

	{#if data.colors.length}
		<RailRow label="Colour palette" class="py-8 sm:py-10">
			<div class="mb-6">
				<h2 class="text-heading-sm font-medium">Colour</h2>
				<p class="mt-2 text-body text-muted-foreground">
					Select any swatch to copy it. Contrast is measured against white and near-black.
				</p>
			</div>
			<ColorPalette colors={data.colors} />
		</RailRow>
	{/if}

	{#if data.typography.length || data.scale.length}
		<RailRow label="Typography" class="py-8 sm:py-10">
			<div class="mb-6">
				<h2 class="text-heading-sm font-medium">Typography</h2>
				<p class="mt-2 text-body text-muted-foreground">
					Families, weights and metrics as measured, with the fallback each one names.
				</p>
			</div>
			<TypographySpecimen roles={data.typography} fonts={data.fonts} scale={data.scale} />
		</RailRow>
	{/if}

	{#if data.spacing.length}
		<RailRow label="Spacing and shape" class="py-8 sm:py-10">
			<h2 class="text-heading-sm font-medium">Spacing and shape</h2>
			<ul class="mt-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
				{#each data.spacing as item (item.key)}
					<li class="panel-card p-4">
						<p class="text-caption text-muted-foreground">{item.key}</p>
						<p class="mt-1 font-mono text-body">{item.value}</p>
					</li>
				{/each}
			</ul>
		</RailRow>
	{/if}

	{#if data.components.length}
		<RailRow label="Components" class="py-8 sm:py-10">
			<h2 class="text-heading-sm font-medium">Components</h2>
			<p class="mt-2 max-w-2xl text-body text-muted-foreground">
				{data.components.length} components, previewed with the colours, radii and spacing recorded
				for each one.
			</p>
			<div class="mt-6">
				<ComponentGallery components={data.components} colors={data.colors} />
			</div>
		</RailRow>
	{/if}

	{#if dos.length || donts.length}
		<RailRow label="Guidelines" class="py-8 sm:py-10">
			<h2 class="text-heading-sm font-medium">Guidelines</h2>
			<div class="mt-6 grid gap-6 lg:grid-cols-2">
				{#if dos.length}
					<div class="panel-card p-5">
						<h3 class="flex items-center gap-2 text-body-lg font-medium">
							<IconCheck size={18} stroke={2} class="text-success" aria-hidden="true" />
							Do
						</h3>
						<ul class="mt-4 flex flex-col gap-3">
							{#each dos as item (item.text)}
								<li class="text-body leading-relaxed text-muted-foreground">{item.text}</li>
							{/each}
						</ul>
					</div>
				{/if}
				{#if donts.length}
					<div class="panel-card p-5">
						<h3 class="flex items-center gap-2 text-body-lg font-medium">
							<IconX size={18} stroke={2} class="text-destructive" aria-hidden="true" />
							Don't
						</h3>
						<ul class="mt-4 flex flex-col gap-3">
							{#each donts as item (item.text)}
								<li class="text-body leading-relaxed text-muted-foreground">{item.text}</li>
							{/each}
						</ul>
					</div>
				{/if}
			</div>
		</RailRow>
	{/if}

	{#if data.similar.length}
		<RailRow label="Similar brands" class="py-8 sm:py-10">
			<h2 class="text-heading-sm font-medium">Similar brands</h2>
			<ul class="mt-6 flex flex-col gap-3">
				{#each data.similar as item (item.business)}
					<li class="panel-card p-4">
						<p class="text-body font-medium">{item.business}</p>
						<p class="mt-1 text-body text-muted-foreground">{item.why}</p>
					</li>
				{/each}
			</ul>
		</RailRow>
	{/if}

	<RailRow label="Provenance" class="py-8 sm:py-10">
		<div class="panel-card flex flex-wrap items-center justify-between gap-4 p-5">
			<div>
				<h2 class="text-body-lg font-medium">Extraction</h2>
				<p class="mt-1 text-caption text-muted-foreground">
					{#if s.extractedAt}Measured {s.extractedAt.slice(0, 10)}{/if}
					{#if s.elementCount}, {s.elementCount.toLocaleString('en-US')} elements inspected{/if}
				</p>
			</div>
			<Button href="/sites/{s.origin}" variant="outline">
				All systems for {s.origin}
				<IconArrowUpRight size={16} stroke={1.75} aria-hidden="true" />
			</Button>
		</div>
	</RailRow>
</RailFrame>
