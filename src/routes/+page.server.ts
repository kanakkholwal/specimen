import { getRecentStyles, getStats } from '$server/db';
import type { PageServerLoad } from './$types';
import { cache } from '$server/cache';

export const load: PageServerLoad = async ({ setHeaders }) => {
	// Both reads are independent, so they must not become a waterfall.
	const [stats, recent] = await Promise.all([getStats(), getRecentStyles(8)]);

	setHeaders({ 'cache-control': cache.listing });
	return { stats, recent };
};
