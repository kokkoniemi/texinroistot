import { error, json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { getBackendHost } from '$lib/server/backend-host';

export const GET: RequestHandler = async ({ url, params, fetch }) => {
	const authorHash = params.authorHash?.trim();
	if (!authorHash) {
		throw error(400, 'authorHash is required');
	}

	const backendHost = getBackendHost();
	const queryString = url.searchParams.toString();
	const targetURL = `${backendHost}/api/authors/${encodeURIComponent(authorHash)}/stories${queryString ? `?${queryString}` : ''}`;
	const res = await fetch(targetURL);

	if (!res.ok) {
		throw error(res.status, `Backend author stories endpoint failed with status ${res.status}`);
	}

	const payload = await res.json();
	return json(payload);
};
