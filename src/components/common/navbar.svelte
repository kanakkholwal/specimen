<script lang="ts">
	import { page } from '$app/state';
	import { IconMenu2, IconSearch, IconX } from '@tabler/icons-svelte';
	import Logo from '$components/common/Logo.svelte';
	import ThemeToggle from '$components/common/ThemeToggle.svelte';
	import { nav } from '$lib/constants';
	import { cn } from '$lib/utils';

	let open = $state(false);

	const isActive = (href: string) => page.url.pathname === href || page.url.pathname.startsWith(`${href}/`);
</script>

<header class="fixed inset-x-0 top-0 z-40 flex justify-center px-3 pt-3 sm:pt-4">
	<nav
		aria-label="Primary"
		class="rail-column flex h-14 items-center justify-between rounded-xl border border-border bg-background/85 px-3 shadow-xs backdrop-blur-xl"
	>
		<a
			href="/"
			class="flex h-10 items-center rounded-md px-2 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
		>
			<Logo />
		</a>

		<ul class="hidden items-center gap-1 md:flex">
			{#each nav as item (item.href)}
				<li>
					<a
						href={item.href}
						aria-current={isActive(item.href) ? 'page' : undefined}
						class={cn(
							'flex h-10 items-center rounded-md px-3 text-body transition-colors',
							'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring',
							isActive(item.href)
								? 'bg-muted font-medium text-foreground'
								: 'text-muted-foreground hover:bg-muted/60 hover:text-foreground'
						)}
					>
						{item.label}
					</a>
				</li>
			{/each}
		</ul>

		<div class="flex items-center gap-1">
			<a
				href="/explore"
				class="hidden h-10 items-center gap-2 rounded-md border border-border px-3 text-body text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring sm:flex"
			>
				<IconSearch size={16} stroke={1.75} aria-hidden="true" />
				Search
			</a>
			<ThemeToggle />
			<button
				type="button"
				onclick={() => (open = !open)}
				aria-expanded={open}
				aria-controls="mobile-nav"
				aria-label={open ? 'Close menu' : 'Open menu'}
				class="inline-flex size-10 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring md:hidden"
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
		<ul class="rail-column flex flex-col gap-1 rounded-xl border border-border bg-popover p-2 shadow-md">
			{#each nav as item (item.href)}
				<li>
					<a
						href={item.href}
						onclick={() => (open = false)}
						aria-current={isActive(item.href) ? 'page' : undefined}
						class={cn(
							'flex h-11 items-center rounded-lg px-3 text-body',
							isActive(item.href)
								? 'bg-muted font-medium text-foreground'
								: 'text-muted-foreground hover:bg-muted/60'
						)}
					>
						{item.label}
					</a>
				</li>
			{/each}
		</ul>
	</div>
{/if}
