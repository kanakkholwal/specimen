import { getFacets, searchStyles } from '$server/db';
import type { PageServerLoad } from './$types';

const PER_PAGE = 24;

export const load: PageServerLoad = async ({ url, platform, setHeaders }) => {
	const q = url.searchParams.get('q') ?? '';
	const theme = url.searchParams.get('theme') ?? '';
	const industry = url.searchParams.get('industry') ?? '';
	const page = Math.max(1, Number(url.searchParams.get('page') ?? '1') || 1);

	const [search, facets] = await Promise.all([
		searchStyles(platform, {
			q,
			theme,
			industry,
			limit: PER_PAGE,
			offset: (page - 1) * PER_PAGE,
		}),
		getFacets(platform),
	]);

	setHeaders({ 'cache-control': 'public, max-age=0, s-maxage=600' });
	return {
		...search,
		facets,
		filters: { q, theme, industry },
		page,
		perPage: PER_PAGE,
	};
};
