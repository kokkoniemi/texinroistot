import { error, json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ url, fetch }) => {
	const res = await fetch("http://localhost:6969/api/stories");
	const stories = await res.json();
	
	return json(stories);
};

