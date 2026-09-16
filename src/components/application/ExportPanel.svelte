<script lang="ts">
	import { untrack } from 'svelte';
	import { IconCheck, IconCopy, IconDownload, IconLoader2 } from '@tabler/icons-svelte';
	import HighlightedCode from '$components/application/HighlightedCode.svelte';
	import { EXPORT_FORMATS } from '$lib/exports';
	import { cn } from '$lib/utils';

	type Props = {
		styleId: string;
		available: string[];
		initial?: { file: string; body: string } | null;
	};

	let { styleId, available, initial = null }: Props = $props();

	const formats = $derived(
		EXPORT_FORMATS.filter((f) =>
			[f.compact, f.extended, f.file].some((n) => n && available.includes(n))
		)
	);

	let activeLabel = $state('');
	let variant = $state<'compact' | 'extended'>('compact');
	let copied = $state(false);
	let loading = $state(false);
	let failed = $state('');

	// Seeded once from the server-rendered export so the first tab needs no fetch.
	const cache = new Map<string, string>();
	untrack(() => {
		if (initial) cache.set(initial.file, initial.body);
	});

	const active = $derived(formats.find((f) => f.label === activeLabel) ?? formats[0]);
	const hasVariants = $derived(Boolean(active?.compact && active?.extended));
	const currentFile = $derived(
		active?.file ?? (variant === 'extended' ? active?.extended : active?.compact) ?? ''
	);

	let body = $state(untrack(() => initial?.body ?? ''));

	$effect(() => {
		const file = currentFile;
		if (!file) return;

		const cached = cache.get(file);
		if (cached !== undefined) {
			body = cached;
			failed = '';
			return;
		}

		let cancelled = false;
		loading = true;
		failed = '';
		fetch(`/api/style/${styleId}/${encodeURIComponent(file)}`)
			.then((r) => {
				if (!r.ok) throw new Error(`${r.status}`);
				return r.text();
			})
			.then((text) => {
				if (cancelled) return;
				cache.set(file, text);
				body = text;
			})
			.catch(() => {
				if (!cancelled) {
					body = '';
					failed = 'That export could not be loaded.';
				}
			})
			.finally(() => {
				if (!cancelled) loading = false;
			});

		return () => {
			cancelled = true;
		};
	});

	async function copy() {
		try {
			await navigator.clipboard.writeText(body);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		} catch {
			failed = 'Copying was blocked. Select the text and copy manually.';
		}
	}

	function download() {
		const blob = new Blob([body], { type: 'text/plain;charset=utf-8' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = currentFile;
		a.click();
		URL.revokeObjectURL(url);
	}

	const lineCount = $derived(body ? body.split('\n').length : 0);
</script>

<div class="panel-card overflow-hidden">
	<div class="flex flex-wrap items-center gap-1 border-b border-border p-2">
		{#each formats as format (format.label)}
			<button
				type="button"
				onclick={() => (activeLabel = format.label)}
				aria-pressed={activeLabel === format.label}
				class={cn(
					'h-10 rounded-md px-3 text-body transition-colors',
					'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring',
					activeLabel === format.label
						? 'bg-muted font-medium text-foreground'
						: 'text-muted-foreground hover:bg-muted/60 hover:text-foreground'
				)}
			>
				{format.label}
			</button>
		{/each}

		<div class="ms-auto flex items-center gap-1">
			{#if hasVariants}
				<div
					role="group"
					aria-label="Detail level"
					class="mr-1 flex items-center rounded-md border border-border p-0.5"
				>
					{#each ['compact', 'extended'] as const as level (level)}
						<button
							type="button"
							onclick={() => (variant = level)}
							aria-pressed={variant === level}
							class={cn(
								'h-8 rounded-sm px-2.5 text-caption capitalize transition-colors',
								'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring',
								variant === level
									? 'bg-muted font-medium text-foreground'
									: 'text-muted-foreground hover:text-foreground'
							)}
						>
							{level}
						</button>
					{/each}
				</div>
			{/if}

			<button
				type="button"
				onclick={copy}
				disabled={!body}
				class="inline-flex h-10 items-center gap-2 rounded-md bg-action px-3 text-body text-action-foreground shadow-xs transition-opacity hover:opacity-90 disabled:pointer-events-none disabled:opacity-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
			>
				{#if copied}
					<IconCheck size={15} stroke={2} aria-hidden="true" />
					Copied
				{:else}
					<IconCopy size={15} stroke={1.75} aria-hidden="true" />
					Copy
				{/if}
			</button>

			<button
				type="button"
				onclick={download}
				disabled={!body}
				aria-label="Download {currentFile}"
				class="inline-flex size-10 items-center justify-center rounded-md border border-border text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:pointer-events-none disabled:opacity-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
			>
				<IconDownload size={16} stroke={1.75} aria-hidden="true" />
			</button>
		</div>
	</div>

	{#if active?.hint}
		<p class="border-b border-border px-4 py-2 text-caption text-muted-foreground">
			{active.hint}
		</p>
	{/if}

	<div class="relative">
		{#if loading}
			<p class="flex items-center gap-2 p-4 text-body text-muted-foreground">
				<IconLoader2 size={16} stroke={1.75} class="animate-spin" aria-hidden="true" />
				Loading {currentFile}
			</p>
		{:else if failed}
			<p class="p-4 text-body text-destructive">{failed}</p>
		{:else}
			<div class="scrollbar-subtle max-h-128 overflow-auto p-4 text-caption leading-relaxed">
				<HighlightedCode code={body} lang={active?.lang ?? 'markdown'} />
			</div>
		{/if}
	</div>

	<div
		class="flex items-center justify-between gap-3 border-t border-border px-4 py-2 text-caption text-muted-foreground"
	>
		<span class="font-mono">{currentFile}</span>
		<span aria-live="polite">
			{#if !loading && body}{lineCount.toLocaleString('en-US')} lines{/if}
		</span>
	</div>
</div>
