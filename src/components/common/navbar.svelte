<script lang="ts">
	import { page } from '$app/state';
	import { IconBrandGithub, IconMenu2, IconX } from '@tabler/icons-svelte';
	import Logo from '$components/common/Logo.svelte';
	import ThemeToggle from '$components/common/ThemeToggle.svelte';
	import { nav, site } from '$lib/constants';
	import { cn } from '$lib/utils';

	let open = $state(false);

	const isActive = (href: string) =>
		page.url.pathname === href || page.url.pathname.startsWith(`${href}/`);

	// Every control in the bar is the same height and radius, so nothing reads as an
	// exception. Icon buttons are square at the same height.
	const control =
		'inline-flex h-10 items-center justify-center rounded-md text-body transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring';
	const quiet = 'text-muted-foreground hover:bg-muted hover:text-foreground';
</script>

<header class="fixed inset-x-0 top-0 z-40 flex justify-center px-3 pt-3 sm:pt-4">
	<nav
		aria-label="Primary"
		class="rail-column grid h-14 grid-cols-[1fr_auto_1fr] items-center gap-2 rounded-xl border border-border bg-background/85 px-2 shadow-xs backdrop-blur-xl sm:px-3"
	>
		<div class="flex justify-start">
			<a href="/" aria-label="{site.name} home" class={cn(control, 'gap-2 px-2', quiet)}>
				<Logo />
			</a>
		</div>

		<ul class="hidden items-center gap-1 md:flex">
			{#each nav as item (item.href)}
				<li>
					<a
						href={item.href}
						aria-current={isActive(item.href) ? 'page' : undefined}
						class={cn(
							control,
							'px-3',
							isActive(item.href) ? 'bg-muted font-medium text-foreground' : quiet
						)}
					>
						{item.label}
					</a>
				</li>
			{/each}
		</ul>
		<div class="md:hidden"></div>

		<div class="flex items-center justify-end gap-1">
			<a
				href={site.repo}
				target="_blank"
				rel="noreferrer noopener"
				aria-label="Source on GitHub, opens in a new tab"
				class={cn(control, 'hidden w-10 sm:inline-flex', quiet)}
			>
				<IconBrandGithub size={18} stroke={1.75} aria-hidden="true" />
			</a>

			<ThemeToggle />

			<button
				type="button"
				onclick={() => (open = !open)}
				aria-expanded={open}
				aria-controls="mobile-nav"
				aria-label={open ? 'Close menu' : 'Open menu'}
				class={cn(control, 'w-10 md:hidden', quiet)}
			>
				{#if open}
					<IconX size={18} stroke={1.75} aria-hidden="true" />
				{:else}
					<IconMenu2 size={18} stroke={1.75} aria-hidden="true" />
				{/if}
			</button>
		</div>
	</nav>
</header>

{#if open}
	<div id="mobile-nav" class="fixed inset-x-0 top-20 z-40 flex justify-center px-3 md:hidden">
		<ul
			class="rail-column flex flex-col gap-1 rounded-xl border border-border bg-popover p-2 shadow-md"
		>
			{#each nav as item (item.href)}
				<li>
					<a
						href={item.href}
						onclick={() => (open = false)}
						aria-current={isActive(item.href) ? 'page' : undefined}
						class={cn(
							'flex h-11 items-center rounded-lg px-3 text-body',
							isActive(item.href) ? 'bg-muted font-medium text-foreground' : quiet
						)}
					>
						{item.label}
					</a>
				</li>
			{/each}
			<li class="mt-1 border-t border-border pt-1">
				<a
					href={site.repo}
					target="_blank"
					rel="noreferrer noopener"
					class={cn('flex h-11 items-center gap-2 rounded-lg px-3 text-body', quiet)}
				>
					<IconBrandGithub size={16} stroke={1.75} aria-hidden="true" />
					Source on GitHub
					<span class="sr-only">, opens in a new tab</span>
				</a>
			</li>
		</ul>
	</div>
{/if}
