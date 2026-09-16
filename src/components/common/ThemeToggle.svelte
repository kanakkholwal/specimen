<script lang="ts">
	import { IconMoon, IconSun } from '@tabler/icons-svelte';
	import { mode, toggleMode } from 'mode-watcher';
	import { cn } from '$lib/utils';

	let { class: className }: { class?: string } = $props();
</script>

<button
	type="button"
	onclick={toggleMode}
	aria-label={mode.current === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'}
	class={cn(
		'inline-flex size-10 items-center justify-center rounded-md text-muted-foreground',
		'transition-colors hover:bg-muted hover:text-foreground',
		'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring',
		className
	)}
>
	{#if mode.current === 'dark'}
		<IconSun size={18} stroke={1.75} aria-hidden="true" />
	{:else}
		<IconMoon size={18} stroke={1.75} aria-hidden="true" />
	{/if}
</button>
