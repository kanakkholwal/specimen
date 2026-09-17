import { error } from '@sveltejs/kit';
import { getCollection } from '$server/db';
import type { PageServerLoad } from './$types';
import { cache } from '$server/cache';

export const load: PageServerLoad = async ({ params, setHeaders }) => {
	const data = await getCollection(params.slug);
	if (!data) error(404, 'No collection with that slug.');
	setHeaders({ 'cache-control': cache.listing });
	return data;
};
