<script lang="ts">
	import { IconBrandGithub, IconMail } from '@tabler/icons-svelte';
	import FooterWordmark from '$components/site/FooterWordmark.svelte';
	import Logo from '$components/common/Logo.svelte';
	import { nav, site } from '$lib/constants';

	const year = new Date().getFullYear();

	const project = [
		{ href: site.repo, label: 'Source on GitHub', external: true },
		{ href: `mailto:${site.support}`, label: 'Contact support', external: false }
	];
</script>

<footer class="flex justify-center px-3 pb-3 sm:px-6 sm:pb-6">
	<div class="rail-column panel-card overflow-hidden rounded-3xl">
		<div class="grid gap-10 p-6 sm:p-10 lg:grid-cols-[minmax(0,1fr)_auto] lg:gap-16">
			<div class="max-w-md">
				<Logo class="text-body-xl" />
				<p class="mt-3 text-body text-muted-foreground">{site.description}</p>
				<p class="mt-4 text-caption text-muted-foreground">
					A <a
						href={site.parent.url}
						class="text-foreground underline underline-offset-4 hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
						>{site.parent.name}</a
					> project.
				</p>
			</div>

			<nav aria-label="Footer" class="grid grid-cols-2 gap-10 sm:gap-16">
				<div>
					<h2 class="label-eyebrow text-muted-foreground">Browse</h2>
					<ul class="mt-4 flex flex-col gap-3">
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

				<div>
					<h2 class="label-eyebrow text-muted-foreground">Project</h2>
					<ul class="mt-4 flex flex-col gap-3">
						{#each project as item (item.href)}
							<li>
								<a
									href={item.href}
									target={item.external ? '_blank' : undefined}
									rel={item.external ? 'noreferrer noopener' : undefined}
									class="inline-flex items-center gap-2 text-body text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
								>
									{#if item.external}
										<IconBrandGithub size={15} stroke={1.75} aria-hidden="true" />
									{:else}
										<IconMail size={15} stroke={1.75} aria-hidden="true" />
									{/if}
									{item.label}
									{#if item.external}<span class="sr-only">, opens in a new tab</span>{/if}
								</a>
							</li>
						{/each}
					</ul>
				</div>
			</nav>
		</div>

		<div class="px-6 pb-2 sm:px-10">
			<FooterWordmark text={site.wordmark} />
		</div>

		<div
			class="flex flex-col gap-2 border-t border-border px-6 py-5 text-caption text-muted-foreground sm:flex-row sm:items-center sm:justify-between sm:px-10"
		>
			<p>{year} {site.parent.name}</p>
			<p>Design data extracted from public product websites.</p>
		</div>
	</div>
</footer>
