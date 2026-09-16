<script lang="ts">
	import { site } from '$lib/constants';
	import { cn } from '$lib/utils';

	let { class: className, mark = true }: { class?: string; mark?: boolean } = $props();
</script>

<span class={cn('inline-flex items-center gap-2', className)}>
	{#if mark}
		<!-- Three descending bars: a type specimen sheet reduced to a mark. -->
		<svg viewBox="0 0 24 24" class="size-6 shrink-0" aria-hidden="true">
			<rect width="24" height="24" rx="7" class="fill-primary" />
			<g class="fill-primary-foreground">
				<rect x="6" y="7" width="12" height="2.5" rx="1.25" />
				<rect x="6" y="11" width="9" height="2.5" rx="1.25" />
				<rect x="6" y="15" width="5" height="2.5" rx="1.25" />
			</g>
		</svg>
	{/if}
	<span class="font-display text-body-lg font-medium tracking-[-0.02em] text-foreground">
		{site.wordmark}
	</span>
</span>
