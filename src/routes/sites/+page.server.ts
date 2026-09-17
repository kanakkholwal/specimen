import { getSites } from '$server/db';
import type { PageServerLoad } from './$types';
import { cache } from '$server/cache';

export const load: PageServerLoad = async ({ setHeaders }) => {
	const data = await getSites(2000);
	setHeaders({ 'cache-control': cache.listing });
	return data;
};
