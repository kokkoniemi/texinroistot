import { error, json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { BACKEND_HOST } from '$env/static/private';

export const GET: RequestHandler = async ({ url, fetch }) => {
	const queryString = url.searchParams.toString();
	const targetURL = `${BACKEND_HOST}/api/villains${queryString ? `?${queryString}` : ''}`;
	const res = await fetch(targetURL);

	if (!res.ok) {
		throw error(res.status, `Backend villains endpoint failed with status ${res.status}`);
	}

	const payload = await res.json();
	return json(payload);
};
