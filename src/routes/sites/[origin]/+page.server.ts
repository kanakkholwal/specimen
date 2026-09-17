import { error } from '@sveltejs/kit';
import { getStylesForOrigin } from '$server/db';
import type { PageServerLoad } from './$types';
import { cache } from '$server/cache';

export const load: PageServerLoad = async ({ params, setHeaders }) => {
	const styles = await getStylesForOrigin(params.origin);
	if (!styles.length) error(404, `Nothing stored for ${params.origin}.`);
	setHeaders({ 'cache-control': cache.listing });
	return { origin: params.origin, styles };
};
