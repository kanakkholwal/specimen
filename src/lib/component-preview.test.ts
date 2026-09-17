/// <reference types="bun" />
import { describe, expect, test } from 'bun:test';
import {
	canvasColor,
	type ComponentKind,
	factChips,
	kindOf,
	parseFacts,
} from './component-preview';

describe('parseFacts', () => {
	test('reads values stated after the property', () => {
		const f = parseFacts('Background #4b72f3, text #ffffff, border-radius 16px, padding 12px 24px');
		expect(f).toMatchObject({
			background: '#4b72f3',
			foreground: '#ffffff',
			radius: 16,
			padY: 12,
			padX: 24,
		});
	});

	test('reads values stated before the property', () => {
		const f = parseFacts('Dark charcoal fill (#131315), white text, 35px border-radius');
		expect(f).toMatchObject({ background: '#131315', foreground: '#ffffff', radius: 35 });
	});

	test('prefers a colour literal over a colour word in the same clause', () => {
		const f = parseFacts('Sticky white bar with 1px bottom hairline (#d0d0d0)');
		expect(f.border).toBe('#d0d0d0');
	});

	test('keeps vertical and horizontal padding apart however they are worded', () => {
		const f = parseFacts('~15px vertical / 24px horizontal padding');
		expect(f).toMatchObject({ padY: 15, padX: 24 });
	});

	test('follows the stated direction rather than the order of the numbers', () => {
		const f = parseFacts('padding of 24px horizontal and 15px vertical');
		expect(f).toMatchObject({ padY: 15, padX: 24 });
	});

	test('ignores state clauses so hover colours never become the resting ones', () => {
		const f = parseFacts('Background #ffffff. Hover: background #000000');
		expect(f.background).toBe('#ffffff');
	});

	test('ignores shadow offsets, which are lengths but not spacing', () => {
		const f = parseFacts('Carries the 8px 8px 0px 0px #f1f1f1 hard offset shadow');
		expect(f).toMatchObject({ padY: null, radius: null, measured: false });
	});

	test('takes the first number of a range', () => {
		expect(parseFacts('border-radius 8px-12px').radius).toBe(8);
	});

	test('reports prose with no measurements as unmeasured', () => {
		expect(parseFacts('A row of partner logos, evenly spaced').measured).toBe(false);
	});
});

describe('kindOf', () => {
	const cases: [string, ComponentKind][] = [
		['Primary CTA Button', 'button'],
		['Top Navigation Bar', 'nav'],
		['Announcement Bar', 'banner'],
		['Input Field', 'input'],
		['Hairline Divider', 'divider'],
		['Hero Display Headline', 'heading'],
		['Feature Card', 'card'],
		['Ghost Text Link', 'link'],
		['Tag / Category Pill', 'pill'],
	];

	for (const [name, expected] of cases) {
		test(`${name} is a ${expected}`, () => {
			expect(kindOf(name, '')).toBe(expected);
		});
	}
});

describe('presentation', () => {
	test('chips omit anything that was not measured', () => {
		const chips = factChips(parseFacts('border-radius 16px'));
		expect(chips).toEqual([{ key: 'Radius', value: '16px' }]);
	});

	test('equal padding is written once', () => {
		expect(factChips(parseFacts('padding 12px')).find((c) => c.key === 'Padding')?.value).toBe(
			'12px',
		);
	});

	test('canvas comes from a named background, not the first colour', () => {
		const colors = [
			{ hex: '#ff0000', name: 'Accent', group: 'brand' },
			{ hex: '#faf2ec', name: 'Parchment Cream', group: 'background' },
		];
		expect(canvasColor(colors)).toBe('#faf2ec');
	});

	test('canvas is absent when nothing names one', () => {
		expect(canvasColor([{ hex: '#ff0000', name: 'Accent', group: 'brand' }])).toBeNull();
	});
});
