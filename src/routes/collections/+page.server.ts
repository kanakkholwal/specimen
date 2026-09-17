import { getCollections } from '$server/db';
import type { PageServerLoad } from './$types';
import { cache } from '$server/cache';

export const load: PageServerLoad = async ({ setHeaders }) => {
	setHeaders({ 'cache-control': cache.listing });
	return { collections: await getCollections() };
};
