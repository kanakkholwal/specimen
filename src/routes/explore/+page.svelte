<script lang="ts">
	import { IconFilter, IconSearch, IconX } from '@tabler/icons-svelte';
	import StyleCardItem from '$components/application/StyleCard.svelte';
	import RailFrame from '$components/site/RailFrame.svelte';
	import RailRow from '$components/site/RailRow.svelte';
	import { Button } from '$components/ui/button';
	import { site } from '$lib/constants';
	import { cn } from '$lib/utils';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const hasFilters = $derived(
		Boolean(data.filters.q || data.filters.theme || data.filters.industry)
	);
	const lastPage = $derived(Math.max(1, Math.ceil(data.total / data.perPage)));

	function pageHref(page: number) {
		const p = new URLSearchParams();
		if (data.filters.q) p.set('q', data.filters.q);
		if (data.filters.theme) p.set('theme', data.filters.theme);
		if (data.filters.industry) p.set('industry', data.filters.industry);
		if (page > 1) p.set('page', String(page));
		const qs = p.toString();
		return qs ? `/explore?${qs}` : '/explore';
	}

	function facetHref(key: 'theme' | 'industry', value: string) {
		const p = new URLSearchParams();
		if (data.filters.q) p.set('q', data.filters.q);
		const current = { theme: data.filters.theme, industry: data.filters.industry };
		for (const k of ['theme', 'industry'] as const) {
			const next = k === key ? (current[k] === value ? '' : value) : current[k];
			if (next) p.set(k, next);
		}
		const qs = p.toString();
		return qs ? `/explore?${qs}` : '/explore';
	}
</script>

<svelte:head>
	<title>Explore design systems | {site.name}</title>
	<meta name="description" content="Search {data.total} extracted design systems by company, industry and theme." />
</svelte:head>

<RailFrame>
	<RailRow divider={false} label="Search" class="pt-28 pb-6 sm:pt-32">
		<h1 class="text-heading-lg font-medium md:text-display">
			Explore
			<span class="text-primary">every system</span>
		</h1>

		<form method="GET" action="/explore" class="mt-6 flex max-w-2xl flex-col gap-3 sm:flex-row">
			<label for="q" class="sr-only">Search design systems</label>
			<div class="relative flex-1">
				<IconSearch
					size={18}
					stroke={1.75}
					aria-hidden="true"
					class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
				/>
				<input
					id="q"
					name="q"
					type="search"
					value={data.filters.q}
					autocomplete="off"
					placeholder="Company, industry or feel"
					class="h-12 w-full rounded-xl border border-input bg-card ps-10 pe-4 text-body outline-none placeholder:text-placeholder focus-visible:border-ring"
				/>
			</div>
			{#if data.filters.theme}<input type="hidden" name="theme" value={data.filters.theme} />{/if}
			{#if data.filters.industry}
				<input type="hidden" name="industry" value={data.filters.industry} />
			{/if}
			<Button type="submit" variant="primary" size="lg" class="h-12">Search</Button>
		</form>
	</RailRow>

	<RailRow label="Filters" class="py-6">
		<div class="flex flex-col gap-4">
			<div class="flex items-center gap-2 text-caption text-muted-foreground">
				<IconFilter size={14} stroke={1.75} aria-hidden="true" />
				<span>Theme</span>
			</div>
			<ul class="flex flex-wrap gap-2">
				{#each data.facets.themes as facet (facet.value)}
					{@const active = data.filters.theme === facet.value}
					<li>
						<a
							href={facetHref('theme', facet.value)}
							aria-current={active ? 'true' : undefined}
							class={cn(
								'inline-flex h-10 items-center gap-2 rounded-md border px-3 text-body transition-colors',
								'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring',
								active
									? 'border-primary bg-primary/10 font-medium text-primary'
									: 'border-border text-muted-foreground hover:bg-muted'
							)}
						>
							{#if active}
								<IconX size={14} stroke={2} aria-hidden="true" />
								<span class="sr-only">Active filter, select to remove:</span>
							{/if}
							{facet.value}
							<span class="text-caption tabular-nums opacity-70">{facet.count}</span>
						</a>
					</li>
				{/each}
			</ul>

			<div class="mt-2 flex items-center gap-2 text-caption text-muted-foreground">
				<IconFilter size={14} stroke={1.75} aria-hidden="true" />
				<span>Industry</span>
			</div>
			<ul class="flex flex-wrap gap-2">
				{#each data.facets.industries as facet (facet.value)}
					{@const active = data.filters.industry === facet.value}
					<li>
						<a
							href={facetHref('industry', facet.value)}
							aria-current={active ? 'true' : undefined}
							class={cn(
								'inline-flex h-10 items-center gap-2 rounded-md border px-3 text-body transition-colors',
								'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring',
								active
									? 'border-primary bg-primary/10 font-medium text-primary'
									: 'border-border text-muted-foreground hover:bg-muted'
							)}
						>
							{#if active}
								<IconX size={14} stroke={2} aria-hidden="true" />
								<span class="sr-only">Active filter, select to remove:</span>
							{/if}
							{facet.value}
							<span class="text-caption tabular-nums opacity-70">{facet.count}</span>
						</a>
					</li>
				{/each}
			</ul>
		</div>
	</RailRow>

	<RailRow label="Results" class="py-8 sm:py-10">
		<div class="flex flex-wrap items-baseline justify-between gap-3">
			<h2 class="text-subheading font-medium">
				{data.total.toLocaleString('en-US')}
				{data.total === 1 ? 'system' : 'systems'}
			</h2>
			{#if hasFilters}
				<a
					href="/explore"
					class="text-body text-muted-foreground underline underline-offset-4 hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
				>
					Clear filters
				</a>
			{/if}
		</div>

		{#if data.results.length}
			<ul class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
				{#each data.results as style (style.id)}
					<li class="flex"><StyleCardItem {style} class="w-full" /></li>
				{/each}
			</ul>

			{#if lastPage > 1}
				<nav aria-label="Pagination" class="mt-10 flex items-center justify-between gap-4">
					{#if data.page > 1}
						<Button href={pageHref(data.page - 1)} variant="outline">Previous</Button>
					{:else}
						<span></span>
					{/if}
					<p class="text-caption text-muted-foreground">Page {data.page} of {lastPage}</p>
					{#if data.page < lastPage}
						<Button href={pageHref(data.page + 1)} variant="outline">Next</Button>
					{:else}
						<span></span>
					{/if}
				</nav>
			{/if}
		{:else}
			<div class="mt-6 panel-card p-8 text-center">
				<p class="text-body-lg font-medium">Nothing matched</p>
				<p class="mt-2 text-body text-muted-foreground">
					{#if data.filters.q}
						No system matches "{data.filters.q}" with the current filters.
					{:else}
						No system matches the current filters.
					{/if}
				</p>
				<Button href="/explore" variant="outline" class="mt-6">Clear filters</Button>
			</div>
		{/if}
	</RailRow>
</RailFrame>
