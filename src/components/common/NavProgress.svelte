<script lang="ts">
	import { navigating } from '$app/state';
	import { prefersReducedMotion } from '$lib/motion';

	// A thin accent bar while a navigation is in flight. Reduced motion gets a static bar
	// rather than the trickle, so nothing animates against the OS setting.
	let width = $state(0);
	let visible = $state(false);

	$effect(() => {
		if (!navigating.to) {
			if (!visible) return;
			width = 100;
			const hide = setTimeout(() => {
				visible = false;
				width = 0;
			}, 200);
			return () => clearTimeout(hide);
		}

		visible = true;
		if (prefersReducedMotion()) {
			width = 60;
			return;
		}

		width = 8;
		const trickle = setInterval(() => {
			const remaining = 90 - width;
			if (remaining > 0) width += remaining * 0.12;
		}, 200);
		return () => clearInterval(trickle);
	});
</script>

{#if visible}
	<div
		role="progressbar"
		aria-label="Loading page"
		aria-valuemin={0}
		aria-valuemax={100}
		aria-valuenow={Math.round(width)}
		class="fixed inset-x-0 top-0 z-50 h-0.5 bg-transparent"
	>
		<div
			class="h-full bg-primary transition-[width] duration-200 ease-craft"
			style:width="{width}%"
		></div>
	</div>
{/if}
