export type ExportFormat = {
	label: string;
	compact?: string;
	extended?: string;
	file?: string;
	lang: string;
	hint: string;
};

export const EXPORT_FORMATS: ExportFormat[] = [
	{
		label: 'DESIGN.md',
		compact: 'DESIGN.compact.md',
		extended: 'DESIGN.extended.md',
		lang: 'markdown',
		hint: 'Drop into a repo so an agent reads the system before it writes code.'
	},
	{
		label: 'CSS Variables',
		compact: 'variables.compact.css',
		extended: 'variables.extended.css',
		lang: 'css',
		hint: 'Custom properties on :root.'
	},
	{
		label: 'Tailwind v4',
		compact: 'theme.compact.css',
		extended: 'theme.extended.css',
		lang: 'css',
		hint: 'An @theme block for Tailwind v4.'
	},
	{
		label: 'Design Tokens',
		compact: 'tokens.compact.json',
		extended: 'tokens.extended.json',
		lang: 'json',
		hint: 'W3C design token format.'
	},
	{
		label: 'Design JSON',
		file: 'design-system.json',
		lang: 'json',
		hint: 'The structured system as one object.'
	}
];

export const EXPORT_FILES: string[] = EXPORT_FORMATS.flatMap((f) =>
	[f.compact, f.extended, f.file].filter((n): n is string => Boolean(n))
);
