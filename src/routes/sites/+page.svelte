<script lang="ts">
	import RailFrame from '$components/site/RailFrame.svelte';
	import RailRow from '$components/site/RailRow.svelte';
	import { site } from '$lib/constants';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const grouped = $derived(
		Object.entries(
			data.sites.reduce<Record<string, typeof data.sites>>((acc, s) => {
				const letter = /^[a-z]/i.test(s.origin) ? s.origin[0].toUpperCase() : '#';
				(acc[letter] ??= []).push(s);
				return acc;
			}, {})
		).sort(([a], [b]) => a.localeCompare(b))
	);
</script>

<svelte:head>
	<title>Sites | {site.name}</title>
	<meta name="description" content="Every product site with an extracted design system." />
</svelte:head>

<RailFrame>
	<RailRow divider={false} label="Sites" class="pt-28 pb-8 sm:pt-32">
		<h1 class="text-heading-lg font-medium md:text-display">
			Sites,
			<span class="text-primary">grouped by origin</span>
		</h1>
		<p class="mt-4 max-w-2xl text-body text-muted-foreground">
			{data.total.toLocaleString('en-US')} origins. Subdomains stay separate from their parent domain,
			so app.example.com and example.com are distinct entries.
		</p>
	</RailRow>

	{#each grouped as [letter, sites] (letter)}
		<RailRow label="Sites starting with {letter}" class="py-6">
			<h2 class="label-eyebrow text-muted-foreground">{letter}</h2>
			<ul class="mt-4 grid gap-2 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
				{#each sites as s (s.origin)}
					<li>
						<a
							href="/sites/{s.origin}"
							class="flex items-center justify-between gap-3 rounded-lg border border-border bg-card px-3 py-2.5 text-body transition-colors hover:bg-muted focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
						>
							<span class="truncate">{s.siteName || s.origin}</span>
							{#if s.styles > 1}
								<span class="shrink-0 text-caption tabular-nums text-muted-foreground">
									{s.styles}
								</span>
							{/if}
						</a>
					</li>
				{/each}
			</ul>
		</RailRow>
	{/each}
</RailFrame>
