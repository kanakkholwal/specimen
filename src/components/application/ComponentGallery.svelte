<script lang="ts">
	import { IconCheck, IconCopy } from '@tabler/icons-svelte';
	import { contrast } from '$lib/color';
	import {
		type ComponentKind,
		KIND_LABELS,
		canvasColor,
		factChips,
		kindOf,
		parseFacts,
		sampleFor
	} from '$lib/component-preview';
	import type { ComponentRow, NamedColor } from '$server/db';
	import { cn } from '$lib/utils';

	type Props = { components: ComponentRow[]; colors: NamedColor[] };
	let { components, colors }: Props = $props();

	let active = $state<ComponentKind | 'all'>('all');
	let copied = $state('');
	let announce = $state('');

	const canvas = $derived(canvasColor(colors));
	const ink = $derived(!canvas || contrast(canvas, '#0a0a0a') >= 4.5 ? '#0a0a0a' : '#ffffff');

	const entries = $derived(
		components.map((c) => ({
			...c,
			kind: kindOf(c.name, c.role),
			facts: parseFacts(c.description)
		}))
	);

	const kinds = $derived(
		[...new Set(entries.map((e) => e.kind))].map((kind) => ({
			kind,
			label: KIND_LABELS[kind],
			count: entries.filter((e) => e.kind === kind).length
		}))
	);

	const shown = $derived(active === 'all' ? entries : entries.filter((e) => e.kind === active));

	async function copy(value: string) {
		try {
			await navigator.clipboard.writeText(value);
			copied = value;
			announce = `Copied ${value}`;
			setTimeout(() => {
				if (copied === value) copied = '';
			}, 1600);
		} catch {
			announce = 'Copying was blocked by the browser.';
		}
	}

	function swatches(facts: ReturnType<typeof parseFacts>) {
		return [
			{ key: 'Background', value: facts.background },
			{ key: 'Text', value: facts.foreground },
			{ key: 'Border', value: facts.border }
		].filter((s): s is { key: string; value: string } => Boolean(s.value));
	}

	// Display sizes run to 80px, which would break out of a preview tile.
	function fit(size: number | null, cap: number) {
		return Math.min(size ?? cap, cap);
	}
</script>

<div class="flex flex-wrap items-center justify-between gap-3">
	<div role="group" aria-label="Component kind" class="flex flex-wrap items-center gap-2">
		{#each [{ kind: 'all' as const, label: 'All', count: entries.length }, ...kinds] as chip (chip.kind)}
			<button
				type="button"
				onclick={() => (active = chip.kind)}
				aria-pressed={active === chip.kind}
				class={cn(
					'inline-flex h-10 cursor-pointer items-center gap-1.5 rounded-full border px-4 text-caption transition-colors',
					'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring',
					active === chip.kind
						? 'border-primary bg-primary text-primary-foreground font-medium'
						: 'border-border text-muted-foreground hover:text-foreground'
				)}
			>
				{chip.label}
				<span class="font-mono text-caption">{chip.count}</span>
			</button>
		{/each}
	</div>
</div>

<ul class="mt-6 grid gap-4 lg:grid-cols-2">
	{#each shown as item (item.name)}
		{@const f = item.facts}
		<li class="panel-card flex flex-col overflow-hidden">
			<div
				aria-hidden="true"
				class="flex min-h-36 items-center justify-center border-b border-border p-6"
				style:background-color={canvas ?? 'var(--muted)'}
				style:color={ink}
			>
				{#if !f.measured}
					<p class="text-caption text-muted-foreground">No measured values</p>
				{:else if item.kind === 'divider'}
					<div
						class="w-full"
						style:height="{Math.max(f.borderWidth, 1)}px"
						style:background-color={f.border ?? f.foreground ?? 'currentColor'}
					></div>
				{:else if item.kind === 'heading'}
					<p
						class="text-center leading-tight"
						style:color={f.foreground}
						style:font-size="{fit(f.fontSize, 32)}px"
						style:font-weight={f.weight}
					>
						{sampleFor(item.kind)}
					</p>
				{:else if item.kind === 'card'}
					<div
						class="w-full max-w-64 p-4"
						style:background-color={f.background}
						style:color={f.foreground}
						style:border="{f.borderWidth || 1}px solid {f.border ?? 'currentColor'}"
						style:border-radius="{f.radius ?? 8}px"
					>
						<p style:font-size="{fit(f.fontSize, 15)}px" style:font-weight={f.weight ?? '600'}>
							{sampleFor(item.kind)}
						</p>
						<span class="mt-2 block h-1.5 w-full rounded-full bg-current opacity-20"></span>
						<span class="mt-1.5 block h-1.5 w-2/3 rounded-full bg-current opacity-20"></span>
					</div>
				{:else if item.kind === 'nav'}
					<div
						class="flex w-full items-center justify-between gap-4"
						style:background-color={f.background}
						style:color={f.foreground}
						style:border-radius="{f.radius ?? 0}px"
						style:padding="{f.padY ?? 12}px {Math.min(f.padX ?? 16, 20)}px"
					>
						<span class="h-5 w-5 rounded-md bg-current"></span>
						<div class="flex gap-4" style:font-size="{fit(f.fontSize, 13)}px">
							<span>Product</span><span>Pricing</span><span>Docs</span>
						</div>
					</div>
				{:else if item.kind === 'banner'}
					<div
						class="w-full text-center"
						style:background-color={f.background}
						style:color={f.foreground}
						style:border-radius="{f.radius ?? 0}px"
						style:padding="{f.padY ?? 10}px {Math.min(f.padX ?? 16, 24)}px"
						style:font-size="{fit(f.fontSize, 13)}px"
					>
						{sampleFor(item.kind)}
					</div>
				{:else if item.kind === 'input'}
					<div
						class="w-full max-w-60 opacity-80"
						style:background-color={f.background}
						style:color={f.foreground}
						style:border="{f.borderWidth || 1}px solid {f.border ?? 'currentColor'}"
						style:border-radius="{f.radius ?? 8}px"
						style:padding="{f.padY ?? 12}px {Math.min(f.padX ?? 16, 20)}px"
						style:font-size="{fit(f.fontSize, 15)}px"
					>
						{sampleFor(item.kind)}
					</div>
				{:else if item.kind === 'link'}
					<span
						class="underline underline-offset-4"
						style:color={f.foreground}
						style:font-size="{fit(f.fontSize, 16)}px"
						style:font-weight={f.weight}
					>
						{sampleFor(item.kind)}
					</span>
				{:else}
					<span
						class="inline-flex items-center"
						style:background-color={f.background}
						style:color={f.foreground}
						style:border={f.border ? `${f.borderWidth || 1}px solid ${f.border}` : undefined}
						style:border-radius="{f.radius ?? 8}px"
						style:padding="{f.padY ?? 12}px {Math.min(f.padX ?? 24, 28)}px"
						style:font-size="{fit(f.fontSize, 16)}px"
						style:font-weight={f.weight}
					>
						{sampleFor(item.kind)}
					</span>
				{/if}
			</div>

			<div class="flex flex-1 flex-col p-5">
				<p class="label-eyebrow uppercase text-muted-foreground">{KIND_LABELS[item.kind]}</p>
				<h3 class="mt-1 text-body-lg font-medium">{item.name}</h3>
				<p class="mt-1 text-caption text-primary">{item.role}</p>

				{#if f.measured}
					<div class="mt-4 flex flex-wrap gap-2">
						{#each swatches(f) as s (s.key)}
							<button
								type="button"
								onclick={() => copy(s.value)}
								title="Copy {s.value}"
								class="inline-flex h-8 cursor-pointer items-center gap-2 rounded-md border border-border pr-2.5 pl-1.5 font-mono text-caption transition-colors hover:border-border-strong focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
							>
								<span
									class="size-4 rounded-sm border border-border"
									style:background-color={s.value}
								></span>
								{s.value}
								{#if copied === s.value}
									<IconCheck size={13} stroke={2} aria-hidden="true" />
								{:else}
									<IconCopy size={13} stroke={1.75} aria-hidden="true" />
								{/if}
							</button>
						{/each}
						{#each factChips(f) as chip (chip.key)}
							<span
								class="inline-flex h-8 items-center gap-1.5 rounded-md bg-muted px-2.5 text-caption text-muted-foreground"
							>
								{chip.key}
								<span class="font-mono text-foreground">{chip.value}</span>
							</span>
						{/each}
					</div>
				{/if}

				<p class="mt-4 text-body leading-relaxed text-muted-foreground">{item.description}</p>
			</div>
		</li>
	{/each}
</ul>

<p class="mt-4 text-caption text-muted-foreground">
	Previews are drawn from the values recorded in each description, in this page's typeface. They
	show the measurements, not a capture of the original component.
</p>

<p class="sr-only" aria-live="polite">{announce}</p>
