<script lang="ts">
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';

	let { class: className, children }: { class?: string; children: Snippet } = $props();
</script>

<span
	class={cn(
		'inline-flex -rotate-2 items-center gap-2 rounded-md border border-border bg-card px-3 py-1.5 text-caption text-muted-foreground shadow-xs',
		className
	)}
>
	{@render children()}
</span>
