<script lang="ts">
	import '@fontsource-variable/google-sans';
	import '@fontsource-variable/inter';
	import '@fontsource-variable/jetbrains-mono';
	import '@fontsource-variable/source-code-pro';
	import '../app.css';
	import { ModeWatcher } from 'mode-watcher';

	let { children } = $props();
</script>

<a
	href="#main"
	class="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50 focus:rounded-md focus:bg-card focus:px-4 focus:py-2 focus:text-body focus:outline-2 focus:outline-offset-2 focus:outline-ring"
>
	Skip to content
</a>

<ModeWatcher />

{@render children()}
