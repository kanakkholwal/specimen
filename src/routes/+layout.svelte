<script lang="ts">
	import { onMount } from 'svelte';
	import { ModeWatcher } from 'mode-watcher';
	import '@fontsource-variable/google-sans';
	import '@fontsource-variable/inter';
	import '@fontsource-variable/jetbrains-mono';
	import '@fontsource-variable/source-code-pro';
	import '../app.css';
	import NavProgress from '$components/common/NavProgress.svelte';

	let { children } = $props();

	onMount(() => {
		const boot = document.getElementById('boot');
		if (!boot) return;
		boot.dataset.done = '';
		const remove = () => boot.remove();
		boot.addEventListener('transitionend', remove, { once: true });
		// Belt and braces: a skipped transition under reduced motion never fires the event.
		setTimeout(remove, 400);
	});
</script>

<a
	href="#main"
	class="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50 focus:rounded-md focus:bg-card focus:px-4 focus:py-2 focus:text-body focus:outline-2 focus:outline-offset-2 focus:outline-ring"
>
	Skip to content
</a>

<ModeWatcher />
<NavProgress />

{@render children()}
