import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import type { Meta } from '$lib/listing/shared';
import type { Villain } from '$lib/types';

type RoistotPageData = {
	villains: Villain[];
	meta: Meta;
	filters: {
		publication: string;
		sort: string;
		q: string;
	};
};

export const load: PageLoad = async ({ fetch, url }) => {
	const queryString = url.searchParams.toString();
	const res = await fetch(`/api/roistot${queryString ? `?${queryString}` : ''}`);

	if (!res.ok) {
		throw error(res.status, `Failed to load villains (${res.status})`);
	}

	return res.json() as Promise<RoistotPageData>;
};
