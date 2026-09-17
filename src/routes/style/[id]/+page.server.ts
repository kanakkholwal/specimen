import { error } from '@sveltejs/kit';
import { getArtifactBody, getArtifactNames, getStyleDetail } from '$server/db';
import type { PageServerLoad } from './$types';
import { cache } from '$server/cache';

const DEFAULT_EXPORT = 'DESIGN.compact.md';

export const load: PageServerLoad = async ({ params, setHeaders }) => {
	// All three reads are independent, so they run together rather than in sequence.
	const [detail, artifacts, body] = await Promise.all([
		getStyleDetail(params.id),
		getArtifactNames(params.id),
		getArtifactBody(params.id, DEFAULT_EXPORT),
	]);

	if (!detail) error(404, 'No design system with that id is stored here.');

	setHeaders({ 'cache-control': cache.record });
	return {
		...detail,
		artifactNames: artifacts.map((a) => a.name),
		initialExport: body === null ? null : { file: DEFAULT_EXPORT, body },
	};
};
