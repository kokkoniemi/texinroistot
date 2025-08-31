import { error, json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { BACKEND_HOST } from '$env/static/private';

export const GET: RequestHandler = async ({ url, fetch }) => {
	const res = await fetch(`${BACKEND_HOST}/api/stories`);
	const stories = await res.json();
	
	return json(stories);
};

