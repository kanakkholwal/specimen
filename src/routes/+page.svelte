<script lang="ts">
	import RailFrame from '$components/site/RailFrame.svelte';
	import RailRow from '$components/site/RailRow.svelte';
	import { site } from '$lib/constants';
</script>

<svelte:head>
	<title>{site.name}</title>
	<meta name="description" content={site.description} />
</svelte:head>

<RailFrame>
	<RailRow divider={false} label="Introduction" class="pt-32 pb-6 sm:pt-36">
		<h1 class="text-heading-sm font-medium sm:text-heading-lg md:text-display lg:text-display-xl">
			Design systems,
			<span class="block text-primary">read from the source</span>
		</h1>
		<p class="mt-6 max-w-2xl text-body text-muted-foreground md:text-body-lg">
			{site.description}
		</p>
	</RailRow>
</RailFrame>
