<script lang="ts">
	import { IconArrowRight, IconSearch, IconSparkles } from '@tabler/icons-svelte';
	import StyleCardItem from '$components/application/StyleCard.svelte';
	import CountUp from '$components/landing/CountUp.svelte';
	import StatStrip from '$components/landing/StatStrip.svelte';
	import BrandPanel from '$components/site/BrandPanel.svelte';
	import RailFrame from '$components/site/RailFrame.svelte';
	import RailRow from '$components/site/RailRow.svelte';
	import TiltedChip from '$components/site/TiltedChip.svelte';
	import { Button } from '$components/ui/button';
	import { site } from '$lib/constants';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	let query = $state('');
</script>

<svelte:head>
	<title>{site.name}: {site.tagline}</title>
	<meta name="description" content={site.description} />
</svelte:head>

<RailFrame>
	<RailRow divider={false} label="Introduction" class="pt-28 pb-10 sm:pt-32 sm:pb-14">
		<div class="grid gap-10 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end lg:gap-16">
			<div class="max-w-3xl">
				<TiltedChip>
					<IconSparkles size={13} stroke={1.75} aria-hidden="true" />
					<CountUp value={data.stats.styles} /> systems, measured not guessed
				</TiltedChip>

				<h1
					class="mt-6 text-heading-sm font-medium sm:text-heading-lg md:text-display lg:text-display-xl"
				>
					Design systems,
					<span class="block text-primary">read from the source</span>
				</h1>

				<p class="mt-6 max-w-2xl text-body text-muted-foreground md:text-body-lg">
					{site.description}
				</p>

				<form action="/explore" method="GET" class="mt-8 flex max-w-xl gap-2">
					<label for="hero-search" class="sr-only">Search design systems</label>
					<div class="relative flex-1">
						<IconSearch
							size={18}
							stroke={1.75}
							aria-hidden="true"
							class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
						/>
						<input
							id="hero-search"
							name="q"
							bind:value={query}
							type="search"
							autocomplete="off"
							placeholder="Search by company, industry or feel"
							class="h-12 w-full rounded-xl border border-input bg-card ps-10 pe-4 text-body outline-none placeholder:text-placeholder focus-visible:border-ring"
						/>
					</div>
					<Button type="submit" variant="primary" size="lg" class="h-12 shrink-0">Search</Button>
				</form>
			</div>

			<div class="lg:pb-2">
				<StatStrip stats={data.stats} />
			</div>
		</div>
	</RailRow>

	<RailRow label="Recently extracted" class="py-10 sm:py-14">
		<div class="flex flex-wrap items-end justify-between gap-4">
			<div>
				<h2 class="text-heading-lg font-medium">Recently extracted</h2>
				<p class="mt-2 text-body text-muted-foreground">
					The newest systems in the archive, newest extraction first.
				</p>
			</div>
			<Button href="/explore" variant="outline">
				Browse all
				<IconArrowRight size={16} stroke={1.75} aria-hidden="true" />
			</Button>
		</div>

		{#if data.recent.length}
			<ul class="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
				{#each data.recent as style (style.id)}
					<li class="flex"><StyleCardItem {style} class="w-full" /></li>
				{/each}
			</ul>
		{:else}
			<p class="mt-8 text-body text-muted-foreground">
				No systems are indexed yet. Run the crawler and import the index.
			</p>
		{/if}
	</RailRow>

	<RailRow label="Call to action" class="py-10 sm:py-14">
		<BrandPanel
			title="Hand a real design system to your agent"
			body="Every entry exports DESIGN.md, CSS variables, a Tailwind v4 theme and W3C design tokens, generated from the measured extraction rather than written by hand."
		>
			{#snippet actions()}
				<Button href="/explore" variant="light" size="lg">
					Start exploring
					<IconArrowRight size={16} stroke={1.75} aria-hidden="true" />
				</Button>
			{/snippet}
		</BrandPanel>
	</RailRow>
</RailFrame>
