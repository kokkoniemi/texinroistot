import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import type { Meta } from '$lib/listing/shared';
import type { Story } from '$lib/types';

type TarinatPageData = {
	stories: Story[];
	meta: Meta;
	filters: {
		publication: string;
		sort: string;
		q: string;
		year: number;
	};
};

export const load: PageLoad = async ({ fetch, url }) => {
	const queryString = url.searchParams.toString();
	const res = await fetch(`/api/tarinat${queryString ? `?${queryString}` : ''}`);

	if (!res.ok) {
		throw error(res.status, `Failed to load stories (${res.status})`);
	}

	return res.json() as Promise<TarinatPageData>;
};
