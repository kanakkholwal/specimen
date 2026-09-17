<script lang="ts">
	import {
		IconArrowUpRight,
		IconLayoutGrid,
		IconMail,
		IconRefresh,
		IconSearch,
		IconWorld
	} from '@tabler/icons-svelte';
	import { page } from '$app/state';
	import RailFrame from '$components/site/RailFrame.svelte';
	import RailRow from '$components/site/RailRow.svelte';
	import { Button } from '$components/ui/button';
	import { site } from '$lib/constants';

	const isMissing = $derived(page.status === 404);
	const isServer = $derived(page.status >= 500);

	const heading = $derived(
		isMissing
			? 'Nothing stored at that address'
			: isServer
				? 'The index did not answer'
				: 'That request was refused'
	);

	const explains = $derived(
		isMissing
			? 'The page may have been renamed, or the address may have a typo in it. Everything stored here is reachable from the three sections below.'
			: isServer
				? 'The page loaded but the index behind it did not respond. Nothing you did caused this, and retrying often works.'
				: (page.error?.message ?? 'The request could not be completed.')
	);

	const routes = [
		{ href: '/explore', label: 'Explore', hint: 'Every stored system, filterable', icon: IconSearch },
		{ href: '/sites', label: 'Sites', hint: 'Grouped by the product they came from', icon: IconWorld },
		{
			href: '/collections',
			label: 'Collections',
			hint: 'Curated sets by style and industry',
			icon: IconLayoutGrid
		}
	];
</script>

<svelte:head>
	<title>{page.status} | {site.name}</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<RailFrame>
	<RailRow divider={false} label="Error" class="pt-28 pb-10 sm:pt-32">
		<div class="grid gap-10 lg:grid-cols-[minmax(0,1fr)_minmax(0,360px)] lg:gap-14">
			<div>
				<p class="label-eyebrow uppercase text-muted-foreground">
					{isMissing ? 'Not found' : isServer ? 'Service error' : 'Request blocked'}
				</p>
				<h1 class="mt-3 text-heading-lg font-medium md:text-display">{heading}</h1>
				<p class="mt-5 max-w-xl text-body-lg leading-relaxed text-muted-foreground">{explains}</p>

				<div class="mt-8 flex flex-wrap gap-3">
					{#if isServer}
						<Button variant="primary" onclick={() => location.reload()}>
							<IconRefresh size={16} stroke={1.75} aria-hidden="true" />
							Try again
						</Button>
						<Button href="mailto:{site.support}" variant="outline">
							<IconMail size={16} stroke={1.75} aria-hidden="true" />
							Report it
						</Button>
					{:else}
						<Button href="/explore" variant="primary">
							Explore systems
							<IconArrowUpRight size={16} stroke={1.75} aria-hidden="true" />
						</Button>
						<Button href="/" variant="outline">Back to home</Button>
					{/if}
				</div>
			</div>

			<div class="panel-card p-6">
				<dl class="flex flex-col gap-4">
					<div>
						<dt class="text-caption text-muted-foreground">Status</dt>
						<dd class="mt-1 font-mono text-subheading">{page.status}</dd>
					</div>
					<div>
						<dt class="text-caption text-muted-foreground">Address</dt>
						<dd class="mt-1 font-mono text-caption break-all text-foreground">
							{page.url.pathname}
						</dd>
					</div>
					{#if page.error?.message}
						<div>
							<dt class="text-caption text-muted-foreground">Reported</dt>
							<dd class="mt-1 text-body text-foreground">{page.error.message}</dd>
						</div>
					{/if}
				</dl>
			</div>
		</div>
	</RailRow>

	<RailRow label="Sections" class="py-8 sm:py-10">
		<h2 class="text-heading-sm font-medium">Where to go instead</h2>
		<ul class="mt-6 grid gap-4 sm:grid-cols-3">
			{#each routes as route (route.href)}
				<li>
					<a
						href={route.href}
						class="panel-card flex h-full flex-col gap-2 p-5 transition-colors hover:border-border-strong focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
					>
						<route.icon size={20} stroke={1.75} class="text-primary" aria-hidden="true" />
						<span class="text-body-lg font-medium">{route.label}</span>
						<span class="text-caption text-muted-foreground">{route.hint}</span>
					</a>
				</li>
			{/each}
		</ul>
	</RailRow>
</RailFrame>
