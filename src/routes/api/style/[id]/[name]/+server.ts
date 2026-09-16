import { error, text } from '@sveltejs/kit';
import { EXPORT_FILES } from '$lib/exports';
import { getArtifactBody } from '$server/db';
import type { RequestHandler } from './$types';

const ALLOWED = new Set<string>(EXPORT_FILES);

export const GET: RequestHandler = async ({ params, platform, setHeaders }) => {
	// Whitelisted so the route can only ever serve a known export file.
	if (!ALLOWED.has(params.name)) error(404, 'Unknown export format.');

	const body = await getArtifactBody(platform, params.id, params.name);
	if (body === null) error(404, 'That export is not stored for this style.');

	setHeaders({ 'cache-control': 'public, max-age=0, s-maxage=86400' });
	return text(body, {
		headers: { 'content-type': 'text/plain; charset=utf-8' },
	});
};
