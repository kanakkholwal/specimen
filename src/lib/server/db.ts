import { error } from '@sveltejs/kit';

export type StyleCard = {
	id: string;
	siteName: string;
	url: string;
	origin: string;
	theme: string | null;
	industry: string | null;
	northStar: string | null;
	colorScheme: string | null;
	thumbnail: string | null;
	colors: string[];
};

export type CorpusStats = {
	styles: number;
	sites: number;
	colors: number;
	fonts: number;
	components: number;
};

export type Facet = { value: string; count: number };

function db(platform: App.Platform | undefined): D1Database {
	if (!platform?.env?.DB) {
		error(503, 'The index is not reachable. Check the D1 binding.');
	}
	return platform.env.DB;
}

const CARD_SELECT = `
	SELECT st.id, COALESCE(st.site_name,'') AS siteName, COALESCE(st.url,'') AS url,
	       COALESCE(si.origin,'') AS origin, st.theme, st.industry, st.north_star AS northStar,
	       st.color_scheme AS colorScheme,
	       (SELECT m.url FROM style_media sm JOIN media m ON m.id = sm.media_id
	         WHERE sm.style_id = st.id AND sm.role = 'thumbnail' LIMIT 1) AS thumbnail,
	       (SELECT group_concat(c.hex) FROM style_colors sc JOIN colors c ON c.id = sc.color_id
	         WHERE sc.style_id = st.id AND sc.origin = 'system'
	         ORDER BY sc.ord LIMIT 6) AS colorList
	FROM styles st LEFT JOIN sites si ON si.id = st.site_id`;

type CardRow = Omit<StyleCard, 'colors'> & { colorList: string | null };

function toCards(rows: CardRow[]): StyleCard[] {
	return rows.map(({ colorList, ...rest }) => ({
		...rest,
		colors: colorList ? colorList.split(',').filter(Boolean).slice(0, 6) : []
	}));
}

export async function getStats(platform: App.Platform | undefined): Promise<CorpusStats> {
	const row = await db(platform)
		.prepare(
			`SELECT (SELECT COUNT(*) FROM styles WHERE has_result = 1) AS styles,
			        (SELECT COUNT(*) FROM sites) AS sites,
			        (SELECT COUNT(*) FROM colors) AS colors,
			        (SELECT COUNT(*) FROM fonts) AS fonts,
			        (SELECT COUNT(*) FROM components) AS components`
		)
		.first<CorpusStats>();
	if (!row) error(503, 'The index returned no stats.');
	return row;
}

export async function getRecentStyles(
	platform: App.Platform | undefined,
	limit = 12
): Promise<StyleCard[]> {
	const { results } = await db(platform)
		.prepare(
			`${CARD_SELECT} WHERE st.has_result = 1 AND st.site_name IS NOT NULL
			 ORDER BY st.extracted_at DESC LIMIT ?`
		)
		.bind(limit)
		.all<CardRow>();
	return toCards(results ?? []);
}

export type SearchArgs = {
	q?: string;
	theme?: string;
	industry?: string;
	limit?: number;
	offset?: number;
};

export async function searchStyles(
	platform: App.Platform | undefined,
	{ q, theme, industry, limit = 24, offset = 0 }: SearchArgs
): Promise<{ results: StyleCard[]; total: number }> {
	const binding = db(platform);
	const where: string[] = ['st.has_result = 1'];
	const args: unknown[] = [];

	if (q?.trim()) {
		// FTS5 rejects bare punctuation, so the term is quoted and wildcarded.
		where.push(`st.id IN (SELECT style_id FROM styles_fts WHERE styles_fts MATCH ?)`);
		args.push(`"${q.trim().replace(/"/g, '')}"*`);
	}
	if (theme) {
		where.push('st.theme = ?');
		args.push(theme);
	}
	if (industry) {
		where.push('st.industry = ?');
		args.push(industry);
	}
	const clause = `WHERE ${where.join(' AND ')}`;

	const [list, count] = await binding.batch<CardRow | { total: number }>([
		binding
			.prepare(`${CARD_SELECT} ${clause} ORDER BY st.site_name LIMIT ? OFFSET ?`)
			.bind(...args, limit, offset),
		binding
			.prepare(`SELECT COUNT(*) AS total FROM styles st ${clause}`)
			.bind(...args)
	]);

	return {
		results: toCards((list.results ?? []) as CardRow[]),
		total: ((count.results ?? [])[0] as { total: number } | undefined)?.total ?? 0
	};
}

export async function getFacets(
	platform: App.Platform | undefined
): Promise<{ themes: Facet[]; industries: Facet[] }> {
	const binding = db(platform);
	const [themes, industries] = await binding.batch<Facet>([
		binding.prepare(
			`SELECT theme AS value, COUNT(*) AS count FROM styles
			 WHERE has_result = 1 AND theme IS NOT NULL AND theme != ''
			 GROUP BY theme ORDER BY count DESC`
		),
		binding.prepare(
			`SELECT industry AS value, COUNT(*) AS count FROM styles
			 WHERE has_result = 1 AND industry IS NOT NULL AND industry != ''
			 GROUP BY industry ORDER BY count DESC LIMIT 24`
		)
	]);
	return { themes: themes.results ?? [], industries: industries.results ?? [] };
}

export type StyleDetail = {
	id: string;
	siteName: string;
	url: string;
	origin: string;
	etld1: string;
	theme: string | null;
	industry: string | null;
	colorScheme: string | null;
	northStar: string | null;
	northStarDetail: string | null;
	description: string | null;
	layout: string | null;
	imagery: string | null;
	density: string | null;
	scaleName: string | null;
	scaleRatio: number | null;
	extractedAt: string | null;
	elementCount: number | null;
};

export type NamedColor = {
	hex: string;
	name: string | null;
	role: string | null;
	group: string | null;
	l: number | null;
	c: number | null;
	h: number | null;
};
export type FontRow = { family: string; source: string; weights: string | null; frequency: number };
export type TypeScaleRow = { role: string; size: number; lineHeight: number; letterSpacing: number };
export type ComponentRow = { name: string; role: string; description: string };
export type GuidelineRow = { kind: string; text: string };
export type SpacingRow = { key: string; value: string };
export type MediaRow = { role: string; kind: string; url: string; width: number | null; height: number | null };
export type ArtifactRow = { name: string; bytes: number };
export type TypographyRole = {
	role: string;
	family: string;
	sizes: string;
	weight: string;
	lineHeight: string;
	substitute: string;
	letterSpacing: string;
	fontFeatures: string;
};

export async function getStyleDetail(platform: App.Platform | undefined, id: string) {
	const binding = db(platform);
	// One batch, so the detail page never becomes a chain of round trips.
	const [
		style,
		colors,
		fonts,
		scale,
		components,
		guidelines,
		spacing,
		media,
		artifacts,
		similar,
		typography
	] = await binding.batch([
			binding
				.prepare(
					`SELECT st.id, COALESCE(st.site_name,'') AS siteName, COALESCE(st.url,'') AS url,
					        COALESCE(si.origin,'') AS origin, COALESCE(si.etld1,'') AS etld1,
					        st.theme, st.industry, st.color_scheme AS colorScheme,
					        st.north_star AS northStar, st.north_star_detail AS northStarDetail,
					        st.description, st.layout, st.imagery, st.density,
					        st.scale_name AS scaleName, st.scale_ratio AS scaleRatio,
					        st.extracted_at AS extractedAt, st.element_count AS elementCount
					 FROM styles st LEFT JOIN sites si ON si.id = st.site_id
					 WHERE st.id = ? AND st.has_result = 1`
				)
				.bind(id),
			binding
				.prepare(
					`SELECT c.hex, sc.name, sc.role, sc.group_name AS "group",
					        c.oklch_l AS l, c.oklch_c AS c, c.oklch_h AS h
					 FROM style_colors sc JOIN colors c ON c.id = sc.color_id
					 WHERE sc.style_id = ? AND sc.origin = 'system' ORDER BY sc.ord`
				)
				.bind(id),
			binding
				.prepare(
					`SELECT f.family, f.source, sf.weights, sf.frequency
					 FROM style_fonts sf JOIN fonts f ON f.id = sf.font_id
					 WHERE sf.style_id = ? ORDER BY sf.frequency DESC`
				)
				.bind(id),
			binding
				.prepare(
					`SELECT role, size, line_height AS lineHeight, letter_spacing AS letterSpacing
					 FROM type_scale WHERE style_id = ? ORDER BY ord`
				)
				.bind(id),
			binding
				.prepare(`SELECT name, role, description FROM components WHERE style_id = ? ORDER BY ord`)
				.bind(id),
			binding
				.prepare(`SELECT kind, text FROM guidelines WHERE style_id = ? ORDER BY kind, ord`)
				.bind(id),
			binding.prepare(`SELECT key, value FROM spacing_map WHERE style_id = ? ORDER BY key`).bind(id),
			binding
				.prepare(
					`SELECT sm.role, m.kind, m.url, m.width, m.height
					 FROM style_media sm JOIN media m ON m.id = sm.media_id
					 WHERE sm.style_id = ?`
				)
				.bind(id),
			binding
				.prepare(`SELECT name, bytes FROM artifacts WHERE style_id = ? ORDER BY name`)
				.bind(id),
			binding
				.prepare(`SELECT business, why FROM similar WHERE style_id = ? ORDER BY ord`)
				.bind(id),
			binding
				.prepare(
					`SELECT role, family, sizes, weight, line_height AS lineHeight,
					        COALESCE(substitute,'') AS substitute,
					        COALESCE(letter_spacing,'') AS letterSpacing,
					        COALESCE(font_features,'') AS fontFeatures
					 FROM typography_roles WHERE style_id = ? ORDER BY ord`
				)
				.bind(id)
		]);

	const detail = ((style.results ?? [])[0] ?? null) as StyleDetail | null;
	if (!detail) return null;

	return {
		style: detail,
		colors: (colors.results ?? []) as NamedColor[],
		fonts: (fonts.results ?? []) as FontRow[],
		scale: (scale.results ?? []) as TypeScaleRow[],
		components: (components.results ?? []) as ComponentRow[],
		guidelines: (guidelines.results ?? []) as GuidelineRow[],
		spacing: (spacing.results ?? []) as SpacingRow[],
		media: (media.results ?? []) as MediaRow[],
		artifacts: (artifacts.results ?? []) as ArtifactRow[],
		similar: (similar.results ?? []) as { business: string; why: string }[],
		typography: (typography.results ?? []) as TypographyRole[]
	};
}

export type SiteRow = { origin: string; etld1: string; siteName: string | null; styles: number };

export async function getSites(platform: App.Platform | undefined, limit = 200, offset = 0) {
	const binding = db(platform);
	const [list, count] = await binding.batch([
		binding
			.prepare(
				`SELECT si.origin, si.etld1, si.site_name AS siteName, COUNT(st.id) AS styles
				 FROM sites si JOIN styles st ON st.site_id = si.id AND st.has_result = 1
				 GROUP BY si.id ORDER BY si.origin LIMIT ? OFFSET ?`
			)
			.bind(limit, offset),
		binding.prepare(`SELECT COUNT(*) AS total FROM sites`)
	]);
	return {
		sites: (list.results ?? []) as SiteRow[],
		total: ((count.results ?? [])[0] as { total: number } | undefined)?.total ?? 0
	};
}

export async function getStylesForOrigin(platform: App.Platform | undefined, origin: string) {
	const { results } = await db(platform)
		.prepare(`${CARD_SELECT} WHERE st.has_result = 1 AND si.origin = ? ORDER BY st.extracted_at DESC`)
		.bind(origin)
		.all<CardRow>();
	return toCards(results ?? []);
}

export type CollectionRow = { slug: string; title: string | null; kind: string | null; styles: number };

export async function getCollections(platform: App.Platform | undefined) {
	const { results } = await db(platform)
		.prepare(
			`SELECT c.slug, c.title, c.kind, COUNT(cs.style_id) AS styles
			 FROM collections c LEFT JOIN collection_styles cs ON cs.collection_id = c.id
			 GROUP BY c.id HAVING styles > 0 ORDER BY styles DESC`
		)
		.all<CollectionRow>();
	return results ?? [];
}

export async function getCollection(platform: App.Platform | undefined, slug: string) {
	const binding = db(platform);
	const [meta, list] = await binding.batch([
		binding.prepare(`SELECT slug, title, kind FROM collections WHERE slug = ?`).bind(slug),
		binding
			.prepare(
				`${CARD_SELECT}
				 JOIN collection_styles cs ON cs.style_id = st.id
				 JOIN collections c ON c.id = cs.collection_id
				 WHERE c.slug = ? AND st.has_result = 1 ORDER BY cs.ord`
			)
			.bind(slug)
	]);
	const info = ((meta.results ?? [])[0] ?? null) as { slug: string; title: string | null; kind: string | null } | null;
	if (!info) return null;
	return { collection: info, styles: toCards((list.results ?? []) as CardRow[]) };
}

// Bodies are stored in ordered chunks because the largest export is 145 KB, past D1's
// statement ceiling. The reader joins them back.
export async function getArtifactBody(
	platform: App.Platform | undefined,
	styleId: string,
	name: string
): Promise<string | null> {
	const { results } = await db(platform)
		.prepare(`SELECT chunk FROM artifact_content WHERE style_id = ? AND name = ? ORDER BY seq`)
		.bind(styleId, name)
		.all<{ chunk: string }>();
	if (!results?.length) return null;
	return results.map((r) => r.chunk).join('');
}

export async function getArtifactNames(platform: App.Platform | undefined, styleId: string) {
	const { results } = await db(platform)
		.prepare(`SELECT name, bytes FROM artifacts WHERE style_id = ? ORDER BY name`)
		.bind(styleId)
		.all<{ name: string; bytes: number }>();
	return results ?? [];
}
