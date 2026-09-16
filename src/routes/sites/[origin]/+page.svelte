<script lang="ts">
	import StyleCardItem from '$components/application/StyleCard.svelte';
	import RailFrame from '$components/site/RailFrame.svelte';
	import RailRow from '$components/site/RailRow.svelte';
	import { site } from '$lib/constants';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<title>{data.origin} | {site.name}</title>
	<meta name="description" content="Design systems extracted from {data.origin}." />
</svelte:head>

<RailFrame>
	<RailRow divider={false} label="Site" class="pt-28 pb-8 sm:pt-32">
		<p class="text-caption text-muted-foreground">
			<a href="/sites" class="hover:text-foreground">Sites</a>
		</p>
		<h1 class="mt-3 font-mono text-heading-lg font-medium md:text-display">{data.origin}</h1>
		<p class="mt-4 text-body text-muted-foreground">
			{data.styles.length}
			{data.styles.length === 1 ? 'extraction' : 'extractions'} stored.
		</p>
	</RailRow>

	<RailRow label="Systems" class="py-8">
		<ul class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
			{#each data.styles as style (style.id)}
				<li class="flex"><StyleCardItem {style} class="w-full" /></li>
			{/each}
		</ul>
	</RailRow>
</RailFrame>
