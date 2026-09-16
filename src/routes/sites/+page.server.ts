import { getSites } from '$server/db';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ platform, setHeaders }) => {
	const data = await getSites(platform, 2000);
	setHeaders({ 'cache-control': 'public, max-age=0, s-maxage=3600' });
	return data;
};
