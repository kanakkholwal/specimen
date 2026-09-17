import { error, text } from '@sveltejs/kit';
import { EXPORT_FILES } from '$lib/exports';
import { getArtifactBody } from '$server/db';
import type { RequestHandler } from './$types';
import { cache } from '$server/cache';

const ALLOWED = new Set<string>(EXPORT_FILES);

export const GET: RequestHandler = async ({ params, setHeaders }) => {
	// Whitelisted so the route can only ever serve a known export file.
	if (!ALLOWED.has(params.name)) error(404, 'Unknown export format.');

	const body = await getArtifactBody(params.id, params.name);
	if (body === null) error(404, 'That export is not stored for this style.');

	setHeaders({ 'cache-control': cache.record });
	return text(body, {
		headers: { 'content-type': 'text/plain; charset=utf-8' },
	});
};
