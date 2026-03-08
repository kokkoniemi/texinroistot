import { error, json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { getBackendHost } from '$lib/server/backend-host';

export const GET: RequestHandler = async ({ params, url, fetch }) => {
	const backendHost = getBackendHost();
	const storyHash = params.storyHash?.trim();
	if (!storyHash) {
		throw error(400, 'storyHash is required');
	}

	const queryString = url.searchParams.toString();
	const targetURL = `${backendHost}/api/stories/${encodeURIComponent(storyHash)}/villains${queryString ? `?${queryString}` : ''}`;
	const res = await fetch(targetURL);

	if (!res.ok) {
		throw error(res.status, `Backend story villains endpoint failed with status ${res.status}`);
	}

	const payload = await res.json();
	return json(payload);
};
