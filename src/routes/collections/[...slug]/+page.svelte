<script lang="ts">
	import StyleCardItem from '$components/application/StyleCard.svelte';
	import RailFrame from '$components/site/RailFrame.svelte';
	import RailRow from '$components/site/RailRow.svelte';
	import { site } from '$lib/constants';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const title = $derived.by(() => {
		if (data.collection.title) return data.collection.title;
		const last = data.collection.slug.split('/').pop() ?? data.collection.slug;
		const spaced = last.replace(/-/g, ' ');
		return spaced.charAt(0).toUpperCase() + spaced.slice(1);
	});
</script>

<svelte:head>
	<title>{title} | {site.name}</title>
	<meta name="description" content="Design systems in the {title} collection." />
</svelte:head>

<RailFrame>
	<RailRow divider={false} label="Collection" class="pt-28 pb-8 sm:pt-32">
		<p class="text-caption text-muted-foreground">
			<a href="/collections" class="hover:text-foreground">Collections</a>
		</p>
		<h1 class="mt-3 text-heading-lg font-medium md:text-display">{title}</h1>
		<p class="mt-4 text-body text-muted-foreground">
			{data.styles.length}
			{data.styles.length === 1 ? 'system' : 'systems'}.
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
