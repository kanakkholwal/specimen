<script lang="ts">
	import CountUp from '$components/landing/CountUp.svelte';
	import type { CorpusStats } from '$server/db';

	let { stats }: { stats: CorpusStats } = $props();

	const items = $derived([
		{ value: stats.styles, label: 'design systems', suffix: '' },
		{ value: stats.sites, label: 'product sites', suffix: '' },
		{ value: stats.colors, label: 'distinct colours', suffix: '+' },
		{ value: stats.components, label: 'components described', suffix: '+' }
	]);
</script>

<dl class="grid grid-cols-2 gap-x-6 gap-y-6 sm:grid-cols-4">
	{#each items as item (item.label)}
		<div>
			<dt class="sr-only">{item.label}</dt>
			<dd class="font-display text-heading-sm font-medium tabular-nums sm:text-heading">
				<CountUp value={item.value} suffix={item.suffix} />
			</dd>
			<p aria-hidden="true" class="mt-1 text-caption text-muted-foreground">{item.label}</p>
		</div>
	{/each}
</dl>
