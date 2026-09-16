<script lang="ts">
	import type { FontRow, TypeScaleRow, TypographyRole } from '$server/db';

	type Props = { roles: TypographyRole[]; fonts: FontRow[]; scale: TypeScaleRow[] };
	let { roles, fonts, scale }: Props = $props();

	let showAll = $state(false);
	const shown = $derived(showAll ? roles : roles.slice(0, 3));

	const SAMPLE = 'Every letter tells a story worth reading';

	function sourceOf(family: string) {
		return fonts.find((f) => f.family.toLowerCase() === family.toLowerCase())?.source ?? '';
	}

	// The extracted family is almost never installed here, so the specimen renders in the
	// stated fallback and says so rather than pretending.
	function stack(role: TypographyRole) {
		const fallback = role.substitute ? `'${role.substitute}', ` : '';
		return `${fallback}var(--font-sans)`;
	}

	function firstSize(sizes: string) {
		const match = sizes.match(/(\d+(?:\.\d+)?)/g);
		if (!match) return 20;
		const nums = match.map(Number).filter((n) => n >= 10 && n <= 96);
		return nums.length ? Math.max(...nums) : 20;
	}
</script>

<div class="panel-card divide-y divide-border overflow-hidden">
	{#each shown as role (role.family + role.role)}
		{@const size = Math.min(firstSize(role.sizes), 44)}
		<article class="p-5 sm:p-6">
			{#if sourceOf(role.family)}
				<p class="label-eyebrow uppercase text-muted-foreground">{sourceOf(role.family)}</p>
			{/if}

			<h3 class="mt-1 font-mono text-subheading font-medium">{role.family}</h3>

			<p
				class="mt-4 truncate text-foreground"
				style:font-family={stack(role)}
				style:font-size="{size}px"
				style:line-height="1.15"
			>
				{SAMPLE}
			</p>
			<p class="mt-1 text-caption text-muted-foreground">
				Rendered in {role.substitute || 'the page sans'}, not the original face.
			</p>

			<dl class="mt-5 grid grid-cols-2 gap-x-6 gap-y-4 sm:grid-cols-4">
				{#each [{ k: 'Weight', v: role.weight }, { k: 'Sizes', v: role.sizes }, { k: 'Line height', v: role.lineHeight }, { k: 'Letter spacing', v: role.letterSpacing }] as field (field.k)}
					{#if field.v}
						<div>
							<dt class="text-caption text-muted-foreground">{field.k}</dt>
							<dd class="mt-0.5 font-mono text-caption text-foreground">{field.v}</dd>
						</div>
					{/if}
				{/each}
			</dl>

			{#if role.substitute}
				<div class="mt-4">
					<p class="text-caption text-muted-foreground">Fallback</p>
					<p class="mt-0.5 text-body text-foreground">{role.substitute}</p>
				</div>
			{/if}

			{#if role.fontFeatures}
				<div class="mt-4">
					<p class="text-caption text-muted-foreground">OpenType features</p>
					<p class="mt-0.5 font-mono text-caption text-foreground">{role.fontFeatures}</p>
				</div>
			{/if}

			{#if role.role}
				<p class="mt-4 text-body leading-relaxed text-muted-foreground">{role.role}</p>
			{/if}
		</article>
	{/each}

	{#if roles.length > 3}
		<button
			type="button"
			onclick={() => (showAll = !showAll)}
			aria-expanded={showAll}
			class="w-full p-4 text-body text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-ring"
		>
			{showAll ? 'Show fewer' : `Show all ${roles.length} fonts`}
		</button>
	{/if}
</div>

{#if scale.length}
	<div class="panel-card mt-6 divide-y divide-border overflow-hidden">
		{#each scale as step (step.role)}
			<div class="flex flex-wrap items-baseline gap-x-6 gap-y-2 p-4 sm:p-5">
				<span
					class="w-28 shrink-0 rounded-md border border-border px-2 py-0.5 text-center text-caption text-muted-foreground"
				>
					{step.role}
				</span>
				<span class="font-mono text-caption text-muted-foreground">
					{step.size}px / {step.lineHeight}{step.letterSpacing
						? ` / ${step.letterSpacing}px`
						: ''}
				</span>
				<span
					class="min-w-0 flex-1 truncate text-foreground"
					style:font-size="{Math.min(step.size, 40)}px"
					style:line-height="1.15"
				>
					{SAMPLE}
				</span>
			</div>
		{/each}
	</div>
{/if}
