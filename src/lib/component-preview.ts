export type ComponentKind =
	| 'button'
	| 'input'
	| 'card'
	| 'pill'
	| 'link'
	| 'divider'
	| 'heading'
	| 'nav'
	| 'banner'
	| 'layout'
	| 'other';

export type ComponentFacts = {
	background: string | null;
	foreground: string | null;
	border: string | null;
	borderWidth: number;
	radius: number | null;
	padX: number | null;
	padY: number | null;
	fontSize: number | null;
	weight: string | null;
	measured: boolean;
};

const COLOR = /#[0-9a-fA-F]{3,8}|rgba?\([^)]*\)|\btransparent\b|\bwhite\b|\bblack\b/i;
const PX = /(\d+(?:\.\d+)?)\s*px/gi;
const NAMED: Record<string, string> = { white: '#ffffff', black: '#000000' };

// Built from char codes because the formatter unescapes them and the dash gate then fires.
const RANGE_DASH = new RegExp(`[${String.fromCharCode(0x2013, 0x2014)}]`, 'g');

// Descriptions state values in prose and in either order ("35px radius", "radius 35px"), so
// each clause is read for the aspects it mentions rather than matched as a whole.
function clausesOf(description: string) {
	return description
		.replace(RANGE_DASH, '-')
		.split(/[,;.]\s+|\n/)
		.map((c) => c.trim())
		.filter((c) => c && !/shadow|hover|active|focus|transition|animat/i.test(c));
}

function colorIn(clause: string): string | null {
	// A literal wins over a colour word: "white bar with 1px hairline (#d0d0d0)".
	const found = clause.match(/#[0-9a-fA-F]{3,8}|rgba?\([^)]*\)/)?.[0] ?? clause.match(COLOR)?.[0];
	if (!found) return null;
	return NAMED[found.toLowerCase()] ?? found;
}

// The gap may not contain a digit, or "15px vertical / 24px horizontal" reads 15 as both.
const BEFORE_WORD = /(\d+(?:\.\d+)?)\s*px[^\d]{0,12}/;

function lengthBefore(clause: string, word: string): number | null {
	const match = clause.match(new RegExp(BEFORE_WORD.source + word, 'i'));
	return match ? Number(match[1]) : null;
}

function lengths(clause: string): number[] {
	return [...clause.matchAll(PX)].map((m) => Number(m[1]));
}

export function parseFacts(description: string): ComponentFacts {
	const facts: ComponentFacts = {
		background: null,
		foreground: null,
		border: null,
		borderWidth: 0,
		radius: null,
		padX: null,
		padY: null,
		fontSize: null,
		weight: null,
		measured: false,
	};

	for (const clause of clausesOf(description)) {
		const px = lengths(clause);
		const color = colorIn(clause);
		const isRadius = /radius|rounded/i.test(clause);
		const isBorder = /border|hairline|stroke|outline|divider|\brule\b/i.test(clause);

		if (isRadius && px.length && facts.radius === null) facts.radius = px[0];

		if (/padding|gutter/i.test(clause) && px.length && facts.padY === null) {
			const vertical = lengthBefore(clause, 'vertical');
			const horizontal = lengthBefore(clause, 'horizontal');
			facts.padY = vertical ?? px[0];
			facts.padX = horizontal ?? px[1] ?? px[0];
		}

		if (isBorder && !isRadius) {
			if (color && !facts.border) facts.border = color;
			if (px.length && !facts.borderWidth) facts.borderWidth = px[0];
		}

		if (color && !isBorder) {
			if (/background|\bfill\b|\bbg\b|canvas|surface|\bbar\b/i.test(clause) && !facts.background) {
				facts.background = color;
			} else if (/\btext\b|colou?r\b|label|link|type/i.test(clause) && !facts.foreground) {
				facts.foreground = color;
			}
		}

		const weight = clause.match(/weight\s*(\d{3})|\b(\d{3})\s*(?:weight|at\s+\d)/i);
		if (weight && !facts.weight) facts.weight = weight[1] ?? weight[2];
		if (px.length && !isRadius && !isBorder && facts.fontSize === null) {
			if (/font|weight|\bat\s+\d|\btext\b|px\s*\/|type|headline|label/i.test(clause)) {
				facts.fontSize = px.find((n) => n >= 10 && n <= 200) ?? null;
			}
		}
	}

	facts.measured = Boolean(
		facts.background || facts.foreground || facts.border || facts.radius || facts.fontSize,
	);
	return facts;
}

const RULES: [ComponentKind, RegExp][] = [
	['nav', /\bnav(igation)?\b|navbar|menu|sidebar|breadcrumb|\btabs?\b/i],
	['banner', /announcement|banner|ticker|marquee|strip|toast|alert|cookie/i],
	['input', /input|field|search|textarea|select|form|checkbox|toggle|switch/i],
	['button', /button|cta\b/i],
	['link', /\blink\b/i],
	['pill', /pill|tag\b|badge|chip|label|eyebrow/i],
	['divider', /divider|hairline|\brule\b|separator/i],
	['heading', /headline|heading|title|display|type|text\b/i],
	['card', /card|tile|panel|surface|box/i],
	['layout', /section|hero|footer|grid|layout|container|band|column/i],
];

export function kindOf(name: string, role: string): ComponentKind {
	const text = `${name} ${role}`;
	for (const [kind, pattern] of RULES) if (pattern.test(text)) return kind;
	return 'other';
}

export const KIND_LABELS: Record<ComponentKind, string> = {
	button: 'Buttons',
	input: 'Inputs',
	card: 'Cards',
	pill: 'Tags',
	link: 'Links',
	divider: 'Dividers',
	heading: 'Type',
	nav: 'Navigation',
	banner: 'Banners',
	layout: 'Layout',
	other: 'Other',
};

const SAMPLES: Record<ComponentKind, string> = {
	button: 'Get started',
	input: 'Search',
	card: 'Card title',
	pill: 'Category',
	link: 'Learn more',
	divider: '',
	heading: 'Design that reads well',
	nav: 'Product',
	banner: 'Now available',
	layout: 'Section',
	other: 'Sample',
};

export function sampleFor(kind: ComponentKind): string {
	return SAMPLES[kind];
}

/** Picks the page surface a component sits on, so previews are not judged against our own card. */
export function canvasColor(colors: { hex: string; name: string | null; group: string | null }[]) {
	const match = colors.find((c) =>
		/background|canvas|base|surface|page/i.test(`${c.name ?? ''} ${c.group ?? ''}`),
	);
	return match?.hex ?? null;
}

function padding(f: ComponentFacts) {
	if (f.padY === null) return null;
	const differs = f.padX !== null && f.padX !== f.padY;
	return differs ? `${f.padY}px ${f.padX}px` : `${f.padY}px`;
}

export function factChips(facts: ComponentFacts) {
	return [
		{ key: 'Radius', value: facts.radius === null ? null : `${facts.radius}px` },
		{ key: 'Padding', value: padding(facts) },
		{ key: 'Size', value: facts.fontSize === null ? null : `${facts.fontSize}px` },
		{ key: 'Weight', value: facts.weight },
	].filter((chip): chip is { key: string; value: string } => Boolean(chip.value));
}
