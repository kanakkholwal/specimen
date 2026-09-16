export const site = {
	name: 'Specimen',
	wordmark: 'specimen',
	tagline: 'Design systems, read from the source',
	description:
		'An archive of design systems extracted from real product websites: colour tokens, type scales, spacing, components and ready-to-copy DESIGN.md, CSS variables and Tailwind themes.',
	url: 'https://specimen.nexonauts.com',
	repo: 'https://github.com/kanakkholwal/specimen',
	support: 'support@nexonauts.com',
	parent: { name: 'Nexonauts', url: 'https://nexonauts.com' },
} as const;

export const nav = [
	{ href: '/explore', label: 'Explore' },
	{ href: '/sites', label: 'Sites' },
	{ href: '/collections', label: 'Collections' },
] as const;

export const exportFormats = [
	{ file: 'DESIGN.compact.md', label: 'DESIGN.md', hint: 'Compact', lang: 'markdown' },
	{ file: 'DESIGN.extended.md', label: 'DESIGN.md', hint: 'Extended', lang: 'markdown' },
	{ file: 'variables.extended.css', label: 'CSS Variables', hint: 'Extended', lang: 'css' },
	{ file: 'theme.extended.css', label: 'Tailwind v4', hint: 'Extended', lang: 'css' },
	{ file: 'tokens.extended.json', label: 'Design Tokens', hint: 'Extended', lang: 'json' },
	{ file: 'design-system.json', label: 'Design JSON', hint: '', lang: 'json' },
] as const;
