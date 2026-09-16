export type ColorFormat = 'hex' | 'rgb' | 'oklch' | 'var';

export function rgbOf(hex: string): [number, number, number] {
	const v = hex.replace('#', '');
	const full = v.length === 3 ? v.split('').map((c) => c + c).join('') : v;
	return [
		Number.parseInt(full.slice(0, 2), 16),
		Number.parseInt(full.slice(2, 4), 16),
		Number.parseInt(full.slice(4, 6), 16)
	];
}

// WCAG 2.x relative luminance.
export function luminance(hex: string): number {
	const channel = (v: number) => {
		const s = v / 255;
		return s <= 0.03928 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4;
	};
	const [r, g, b] = rgbOf(hex);
	return 0.2126 * channel(r) + 0.7152 * channel(g) + 0.0722 * channel(b);
}

export function contrast(a: string, b: string): number {
	const [x, y] = [luminance(a), luminance(b)];
	const [hi, lo] = x > y ? [x, y] : [y, x];
	return (hi + 0.05) / (lo + 0.05);
}

export function slug(name: string): string {
	return name
		.toLowerCase()
		.replace(/'/g, '')
		.replace(/\s+/g, '-')
		.replace(/[^a-z0-9-]/g, '')
		.replace(/-+/g, '-')
		.replace(/^-|-$/g, '');
}

export function formatColor(
	format: ColorFormat,
	color: { hex: string; name: string | null; l: number | null; c: number | null; h: number | null }
): string {
	switch (format) {
		case 'rgb': {
			const [r, g, b] = rgbOf(color.hex);
			return `rgb(${r} ${g} ${b})`;
		}
		case 'oklch':
			if (color.l === null || color.c === null || color.h === null) return color.hex;
			return `oklch(${(color.l * 100).toFixed(1)}% ${color.c.toFixed(3)} ${color.h.toFixed(1)})`;
		case 'var':
			return `var(--color-${slug(color.name ?? color.hex)})`;
		default:
			return color.hex;
	}
}
