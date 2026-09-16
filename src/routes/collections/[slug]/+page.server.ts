import { error } from '@sveltejs/kit';
import { getCollection } from '$server/db';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, platform, setHeaders }) => {
	const data = await getCollection(platform, decodeURIComponent(params.slug));
	if (!data) error(404, 'No collection with that slug.');
	setHeaders({ 'cache-control': 'public, max-age=0, s-maxage=3600' });
	return data;
};
