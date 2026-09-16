export const faqs = [
	{
		q: 'Where does this data come from?',
		a: 'Each entry records the design decisions a public product site actually uses: its colours, type, spacing and components. Every record stores the date it was measured and the number of elements inspected, so you can see how current it is.',
	},
	{
		q: 'How many design systems are here?',
		a: 'One thousand two hundred and ninety, and the collection grows as new ones are added. Each carries its own measurement date rather than a single freshness claim for the whole set.',
	},
	{
		q: 'Are the exports reliable?',
		a: 'They are generated from the stored measurements rather than written by hand, and checked by tests that fail on the first differing line. What you copy is what was measured.',
	},
	{
		q: 'What is the difference between compact and extended?',
		a: 'Compact is the short reference: colours, type, spacing, components and guidelines. Extended adds token tables with variable names, the measured spacing scale, shadows, and a Quick Start holding ready-to-paste CSS custom properties and a Tailwind v4 theme.',
	},
	{
		q: 'Can I use these in my own project?',
		a: 'Treat them as a reference, not a licence. Colour values, spacing and type scales are design decisions visible to anyone who opens the site, but typefaces, logos and imagery belong to their owners. Check the original site before reusing anything beyond the tokens.',
	},
	{
		q: 'Why does a specimen render in a different typeface?',
		a: 'Because the original face is usually licensed, not free to redistribute. Each entry names the fallback it was measured against and says which one you are looking at, rather than showing a substitute as though it were the real thing.',
	},
];
