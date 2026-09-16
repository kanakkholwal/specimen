export const faqs = [
	{
		q: 'Where does this data come from?',
		a: 'Each entry is an extraction of a public product website: the colours, type, spacing and component descriptions its pages actually use. Every record stores the date it was measured and the number of elements inspected, so you can see how current it is.'
	},
	{
		q: 'How many design systems are here?',
		a: 'One thousand two hundred and ninety. That figure comes from crawling every style page and then following the related-style graph until it closed, rather than from a headline number. Nothing was found outside it.'
	},
	{
		q: 'Are the exports the same as the originals?',
		a: 'Byte for byte. The generators were ported and are checked against the source tool by golden tests that fail on the first differing line, so a DESIGN.md copied here is identical to one copied at the source.'
	},
	{
		q: 'What is the difference between compact and extended?',
		a: 'Compact is the short reference: colours, type, spacing, components and guidelines. Extended adds token tables with variable names, the measured spacing scale, shadows, and a Quick Start holding ready-to-paste CSS custom properties and a Tailwind v4 theme.'
	},
	{
		q: 'Can I use these in my own project?',
		a: 'Treat them as a reference, not a licence. Colour values, spacing and type scales are design decisions visible to anyone who opens the site, but typefaces, logos and imagery belong to their owners. Check the original site before reusing anything beyond the tokens.'
	},
	{
		q: 'How often does it update?',
		a: 'When the crawler is run. There is no automatic schedule, so each entry shows its own extraction date rather than implying the whole archive is fresh.'
	}
];
