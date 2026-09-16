import type { ThemeRegistrationRaw } from 'shiki/core';

// Shiki bundles no JetBrains theme, so these carry the Darcula and IntelliJ Light token
// colours directly. Backgrounds stay transparent: the card behind them owns the surface.
const darculaColors = {
	fg: '#a9b7c6',
	comment: '#808080',
	keyword: '#cc7832',
	string: '#6a8759',
	number: '#6897bb',
	fn: '#ffc66d',
	constant: '#9876aa',
	tag: '#e8bf6a',
	attribute: '#bababa',
	heading: '#ffc66d',
	link: '#287bde'
};

const intellijColors = {
	fg: '#080808',
	comment: '#8c8c8c',
	keyword: '#0033b3',
	string: '#067d17',
	number: '#1750eb',
	fn: '#00627a',
	constant: '#871094',
	tag: '#0033b3',
	attribute: '#174ad4',
	heading: '#00627a',
	link: '#1750eb'
};

function build(name: string, type: 'light' | 'dark', c: typeof darculaColors): ThemeRegistrationRaw {
	return {
		name,
		type,
		colors: { 'editor.background': '#00000000', 'editor.foreground': c.fg },
		settings: [
			{ settings: { foreground: c.fg } },
			{ scope: ['comment', 'punctuation.definition.comment'], settings: { foreground: c.comment } },
			{
				scope: ['keyword', 'storage', 'storage.type', 'keyword.control', 'entity.name.tag'],
				settings: { foreground: c.keyword }
			},
			{
				scope: ['string', 'string.quoted', 'markup.inline.raw', 'markup.raw'],
				settings: { foreground: c.string }
			},
			{
				scope: ['constant.numeric', 'keyword.other.unit', 'constant.other.color'],
				settings: { foreground: c.number }
			},
			{
				scope: ['entity.name.function', 'support.function', 'meta.function-call'],
				settings: { foreground: c.fn }
			},
			{
				scope: ['constant.language', 'variable.other.constant', 'support.constant'],
				settings: { foreground: c.constant }
			},
			{
				scope: [
					'support.type.property-name',
					'meta.object-literal.key',
					'entity.other.attribute-name'
				],
				settings: { foreground: c.attribute }
			},
			{
				scope: ['entity.other.attribute-name.class', 'entity.other.attribute-name.id'],
				settings: { foreground: c.tag }
			},
			{ scope: ['markup.heading', 'entity.name.section'], settings: { foreground: c.heading, fontStyle: 'bold' } },
			{ scope: ['markup.bold'], settings: { fontStyle: 'bold' } },
			{ scope: ['markup.italic'], settings: { fontStyle: 'italic' } },
			{ scope: ['markup.underline.link', 'string.other.link'], settings: { foreground: c.link } },
			{ scope: ['punctuation', 'meta.brace'], settings: { foreground: c.fg } }
		]
	};
}

export const darcula = build('jetbrains-darcula', 'dark', darculaColors);
export const intellijLight = build('jetbrains-light', 'light', intellijColors);
