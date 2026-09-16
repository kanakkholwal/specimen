<script lang="ts">
	import { IconArrowRight } from '@tabler/icons-svelte';
	import RailFrame from '$components/site/RailFrame.svelte';
	import RailRow from '$components/site/RailRow.svelte';
	import { site } from '$lib/constants';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	function label(slug: string) {
		const last = slug.split('/').pop() ?? slug;
		const spaced = last.replace(/-/g, ' ');
		return spaced.charAt(0).toUpperCase() + spaced.slice(1);
	}
</script>

<svelte:head>
	<title>Collections | {site.name}</title>
	<meta name="description" content="Curated groupings of extracted design systems." />
</svelte:head>

<RailFrame>
	<RailRow divider={false} label="Collections" class="pt-28 pb-8 sm:pt-32">
		<h1 class="text-heading-lg font-medium md:text-display">
			Collections,
			<span class="text-primary">by category</span>
		</h1>
		<p class="mt-4 max-w-2xl text-body text-muted-foreground">
			Groupings captured alongside the systems themselves.
		</p>
	</RailRow>

	<RailRow label="All collections" class="py-8">
		{#if data.collections.length}
			<ul class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
				{#each data.collections as c (c.slug)}
					<li>
						<a
							href="/collections/{c.slug}"
							class="panel-card flex items-center justify-between gap-4 p-5 transition-shadow hover:shadow-lg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
						>
							<span>
								<span class="block text-body-lg font-medium">{c.title || label(c.slug)}</span>
								<span class="mt-1 block text-caption text-muted-foreground">
									{c.styles}
									{c.styles === 1 ? 'system' : 'systems'}
								</span>
							</span>
							<IconArrowRight
								size={18}
								stroke={1.75}
								aria-hidden="true"
								class="shrink-0 text-muted-foreground"
							/>
						</a>
					</li>
				{/each}
			</ul>
		{:else}
			<p class="text-body text-muted-foreground">No collections are stored yet.</p>
		{/if}
	</RailRow>
</RailFrame>
