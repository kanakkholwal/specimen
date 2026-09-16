<script lang="ts">
	import FooterWordmark from '$components/site/FooterWordmark.svelte';
	import { nav, site } from '$lib/constants';

	const year = new Date().getFullYear();
</script>

<footer class="flex justify-center px-3 pb-3 sm:px-6 sm:pb-6">
	<div class="rail-column panel-card overflow-hidden rounded-3xl">
		<div class="flex flex-col gap-8 p-6 sm:p-10 lg:flex-row lg:justify-between">
			<div class="max-w-md">
				<p class="font-display text-body-lg font-medium">{site.name}</p>
				<p class="mt-2 text-body text-muted-foreground">{site.description}</p>
			</div>

			<nav aria-label="Footer" class="flex gap-12">
				<div>
					<p class="label-eyebrow text-muted-foreground">Browse</p>
					<ul class="mt-3 flex flex-col gap-2">
						{#each nav as item (item.href)}
							<li>
								<a
									href={item.href}
									class="text-body text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
								>
									{item.label}
								</a>
							</li>
						{/each}
					</ul>
				</div>
			</nav>
		</div>

		<div class="px-6 sm:px-10">
			<FooterWordmark text="design.supply" />
		</div>

		<div
			class="flex flex-col gap-2 border-t border-border px-6 py-5 text-caption text-muted-foreground sm:flex-row sm:items-center sm:justify-between sm:px-10"
		>
			<p>{year} {site.name}</p>
			<p>Design data extracted from public product websites.</p>
		</div>
	</div>
</footer>
