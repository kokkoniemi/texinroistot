import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch, url }) => {
	const queryString = url.searchParams.toString();
	const res = await fetch(`/api/tarinat${queryString ? `?${queryString}` : ''}`);

	if (!res.ok) {
		throw error(res.status, `Failed to load stories (${res.status})`);
	}

	return res.json();
};
