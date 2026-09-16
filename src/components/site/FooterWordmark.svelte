<script lang="ts">
	// SVG text scales with the viewBox, so the wordmark stays off the type scale entirely.
	let { text }: { text: string } = $props();

	let x = $state(50);
	let y = $state(50);
	let on = $state(false);

	function track(event: PointerEvent) {
		if (event.pointerType !== 'mouse') return;
		const rect = (event.currentTarget as HTMLElement).getBoundingClientRect();
		x = ((event.clientX - rect.left) / rect.width) * 100;
		y = ((event.clientY - rect.top) / rect.height) * 100;
		on = true;
	}

	const mask = $derived(`radial-gradient(circle 10rem at ${x}% ${y}%, black, transparent)`);
</script>

<div
	aria-hidden="true"
	class="relative select-none"
	onpointermove={track}
	onpointerleave={() => (on = false)}
>
	<svg viewBox="0 0 1000 220" class="block h-auto w-full fill-muted">
		<text
			x="500"
			y="205"
			text-anchor="middle"
			textLength="990"
			lengthAdjust="spacing"
			font-size="270"
			font-weight="500"
			style="font-family: var(--font-heading)">{text}</text
		>
	</svg>

	<div
		class="absolute inset-0 transition-opacity duration-300"
		style:opacity={on ? 1 : 0}
		style:mask-image={mask}
		style:-webkit-mask-image={mask}
	>
		<svg viewBox="0 0 1000 220" class="block h-auto w-full fill-primary">
			<text
				x="500"
				y="205"
				text-anchor="middle"
				textLength="990"
				lengthAdjust="spacing"
				font-size="270"
				font-weight="500"
				style="font-family: var(--font-heading)">{text}</text
			>
		</svg>
	</div>
</div>
