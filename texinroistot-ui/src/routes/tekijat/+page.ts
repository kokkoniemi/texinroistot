import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import type { Meta } from '$lib/listing/shared';
import type { ListedAuthor } from '$lib/types';

type TekijatPageData = {
	authors: ListedAuthor[];
	meta: Meta;
	filters: {
		type: string;
		sort: string;
		q: string;
	};
};

export const load: PageLoad = async ({ fetch, url }) => {
	const queryString = url.searchParams.toString();
	const res = await fetch(`/api/tekijat${queryString ? `?${queryString}` : ''}`);

	if (!res.ok) {
		throw error(res.status, `Failed to load authors (${res.status})`);
	}

	return res.json() as Promise<TekijatPageData>;
};
