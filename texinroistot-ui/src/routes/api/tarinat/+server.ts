import { error, json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { BACKEND_HOST } from '$env/static/private';

export const GET: RequestHandler = async ({ url, fetch }) => {
	const res = await fetch(`${BACKEND_HOST}/api/stories`);
	const stories = await res.json();
	
	const transform = ({stories}) => ({
		stories: stories.map(({orderNumber, publications, writtenBy, drawnBy}) => ({
			orderNumber,
			title: (publications[0]?.title) ?? "",
			writtenBy,
			drawnBy
		})
	)});

	return json(transform(stories));
};

