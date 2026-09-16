<script lang="ts">
	import { IconMoon, IconPhotoOff, IconSun } from '@tabler/icons-svelte';
	import type { StyleCard } from '$server/db';
	import { cn } from '$lib/utils';

	let { style, class: className }: { style: StyleCard; class?: string } = $props();

	const theme = $derived(style.colorScheme ?? style.theme ?? '');
</script>

<article
	class={cn(
		'group panel-card flex flex-col overflow-hidden transition-shadow hover:shadow-lg',
		className
	)}
>
	<a
		href="/style/{style.id}"
		class="flex flex-col focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
	>
		<div class="relative aspect-[16/10] overflow-hidden border-b border-border bg-muted">
			{#if style.thumbnail}
				<img
					src={style.thumbnail}
					alt="Screenshot of {style.siteName}"
					loading="lazy"
					decoding="async"
					class="size-full object-cover object-top"
				/>
			{:else}
				<div class="flex size-full items-center justify-center text-muted-foreground">
					<IconPhotoOff size={20} stroke={1.5} aria-hidden="true" />
					<span class="sr-only">No screenshot stored</span>
				</div>
			{/if}
		</div>

		<div class="flex flex-1 flex-col gap-2 p-4">
			<div class="flex items-start justify-between gap-3">
				<h3 class="text-body-lg font-medium">{style.siteName}</h3>
				{#if theme}
					<span
						class="inline-flex shrink-0 items-center gap-1 rounded-full border border-border px-2 py-0.5 text-caption text-muted-foreground"
					>
						{#if theme === 'dark'}
							<IconMoon size={12} stroke={1.75} aria-hidden="true" />
						{:else}
							<IconSun size={12} stroke={1.75} aria-hidden="true" />
						{/if}
						{theme}
					</span>
				{/if}
			</div>

			<p class="text-caption text-muted-foreground">{style.origin}</p>

			{#if style.northStar}
				<p class="line-clamp-2 text-body text-muted-foreground">{style.northStar}</p>
			{/if}

			{#if style.colors.length}
				<ul class="mt-auto flex flex-wrap gap-1.5 pt-2">
					{#each style.colors as hex (hex)}
						<li
							class="size-5 rounded-xs border border-border-strong"
							style:background-color={hex}
							title={hex}
						>
							<span class="sr-only">{hex}</span>
						</li>
					{/each}
				</ul>
			{/if}
		</div>
	</a>
</article>
