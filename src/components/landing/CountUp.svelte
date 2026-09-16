<script lang="ts">
	import { untrack } from 'svelte';
	import { prefersReducedMotion } from '$lib/motion';

	let { value, suffix = '', duration = 900 }: { value: number; suffix?: string; duration?: number } =
		$props();

	// Server render and the reduced-motion path both show the final number immediately.
	// untrack is deliberate: the animation owns `shown` after the first paint.
	let shown = $state(untrack(() => value));
	let node: HTMLSpanElement | null = $state(null);

	const format = (n: number) => n.toLocaleString('en-US');

	$effect(() => {
		if (!node || prefersReducedMotion()) return;

		let frame = 0;
		let stop = false;
		const observer = new IntersectionObserver(
			(entries) => {
				if (!entries[0]?.isIntersecting || stop) return;
				stop = true;
				observer.disconnect();
				const start = performance.now();
				const tick = (now: number) => {
					const t = Math.min((now - start) / duration, 1);
					// easeOutCubic, so the number settles rather than stopping dead.
					shown = Math.round(value * (1 - (1 - t) ** 3));
					if (t < 1) frame = requestAnimationFrame(tick);
				};
				shown = 0;
				frame = requestAnimationFrame(tick);
			},
			{ threshold: 0.4 }
		);
		observer.observe(node);
		return () => {
			stop = true;
			observer.disconnect();
			cancelAnimationFrame(frame);
		};
	});
</script>

<span bind:this={node}>{format(shown)}{suffix}</span>
