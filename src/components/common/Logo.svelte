<script lang="ts">
	import { cn } from '$lib/utils';
	import { site } from '$lib/constants';

	let { class: className }: { class?: string } = $props();
</script>

<span class={cn('font-display text-body-lg font-medium tracking-[-0.02em]', className)}>
	<span class="text-foreground">design</span><span class="text-primary">.supply</span>
	<span class="sr-only">{site.tagline}</span>
</span>
