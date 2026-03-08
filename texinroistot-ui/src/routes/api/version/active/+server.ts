import { error, json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { getBackendHost } from '$lib/server/backend-host';

export const GET: RequestHandler = async ({ fetch }) => {
	const res = await fetch(`${getBackendHost()}/api/version/active`);

	if (!res.ok) {
		throw error(res.status, `Backend active version endpoint failed with status ${res.status}`);
	}

	const payload = await res.json();
	return json(payload);
};
