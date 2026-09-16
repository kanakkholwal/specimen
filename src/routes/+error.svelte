<script lang="ts">
	import { page } from '$app/state';
	import RailFrame from '$components/site/RailFrame.svelte';
	import RailRow from '$components/site/RailRow.svelte';
	import { Button } from '$components/ui/button';
	import { site } from '$lib/constants';

	const isMissing = $derived(page.status === 404);
</script>

<svelte:head>
	<title>{page.status} | {site.name}</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<RailFrame>
	<RailRow divider={false} label="Error" class="pt-28 pb-16 sm:pt-32">
		<div class="panel-card mx-auto max-w-xl p-8 text-center sm:p-10">
			<p class="font-mono text-caption text-muted-foreground">{page.status}</p>
			<h1 class="mt-3 text-heading-lg font-medium">
				{#if isMissing}
					Nothing stored
					<span class="block text-primary">at that address</span>
				{:else}
					Something went wrong
					<span class="block text-primary">on our side</span>
				{/if}
			</h1>
			<p class="mt-4 text-body text-muted-foreground">
				{page.error?.message ?? 'The page could not be loaded.'}
			</p>
			<div class="mt-8 flex flex-wrap justify-center gap-3">
				<Button href="/" variant="primary">Back to home</Button>
				<Button href="/explore" variant="outline">Explore systems</Button>
			</div>
		</div>
	</RailRow>
</RailFrame>
