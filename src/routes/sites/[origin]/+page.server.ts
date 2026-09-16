import { error } from '@sveltejs/kit';
import { getStylesForOrigin } from '$server/db';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, platform, setHeaders }) => {
	const styles = await getStylesForOrigin(platform, params.origin);
	if (!styles.length) error(404, `Nothing stored for ${params.origin}.`);
	setHeaders({ 'cache-control': 'public, max-age=0, s-maxage=3600' });
	return { origin: params.origin, styles };
};
