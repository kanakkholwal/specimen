import { error } from '@sveltejs/kit';
import { getArtifactBody, getArtifactNames, getStyleDetail } from '$server/db';
import type { PageServerLoad } from './$types';

const DEFAULT_EXPORT = 'DESIGN.compact.md';

export const load: PageServerLoad = async ({ params, platform, setHeaders }) => {
	// All three reads are independent, so they run together rather than in sequence.
	const [detail, artifacts, body] = await Promise.all([
		getStyleDetail(platform, params.id),
		getArtifactNames(platform, params.id),
		getArtifactBody(platform, params.id, DEFAULT_EXPORT)
	]);

	if (!detail) error(404, 'No design system with that id is stored here.');

	setHeaders({ 'cache-control': 'public, max-age=0, s-maxage=86400' });
	return {
		...detail,
		artifactNames: artifacts.map((a) => a.name),
		initialExport: body === null ? null : { file: DEFAULT_EXPORT, body }
	};
};
