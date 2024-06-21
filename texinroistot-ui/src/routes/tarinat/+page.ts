import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch }) => {
	const res = await fetch("/api/tarinat");
	const stories = res.json();
	
	return stories;
};

