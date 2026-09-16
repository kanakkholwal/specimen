import { getRecentStyles, getStats } from '$server/db';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ platform, setHeaders }) => {
	// Both reads are independent, so they must not become a waterfall.
	const [stats, recent] = await Promise.all([getStats(platform), getRecentStyles(platform, 8)]);

	setHeaders({ 'cache-control': 'public, max-age=0, s-maxage=3600' });
	return { stats, recent };
};
