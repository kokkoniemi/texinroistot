import { redirect, type Handle } from '@sveltejs/kit';
import {
	hasUnpublishedAccess,
	isUnpublishedModeEnabled,
	UNPUBLISHED_ACCESS_COOKIE
} from '$lib/server/unpublished-gate';

const UNPUBLISHED_ROUTE = '/julkaisematon';
const UNPUBLISHED_ALLOWED_API_PATHS = new Set(['/api/login', '/api/logout', '/api/me']);

function isPublicPath(pathname: string): boolean {
	return (
		pathname === UNPUBLISHED_ROUTE ||
		pathname.startsWith('/_app/') ||
		pathname === '/favicon.png' ||
		pathname === '/favicon.ico' ||
		pathname === '/robots.txt' ||
		pathname === '/manifest.webmanifest'
	);
}

function targetWithQuery(url: URL): string {
	return `${url.pathname}${url.search}`;
}

export const handle: Handle = async ({ event, resolve }) => {
	if (!isUnpublishedModeEnabled()) {
		return resolve(event);
	}

	if (isPublicPath(event.url.pathname)) {
		return resolve(event);
	}

	const hasAccess = hasUnpublishedAccess(event.cookies.get(UNPUBLISHED_ACCESS_COOKIE));
	if (hasAccess) {
		return resolve(event);
	}

	if (UNPUBLISHED_ALLOWED_API_PATHS.has(event.url.pathname)) {
		return resolve(event);
	}

	if (event.url.pathname.startsWith('/api/')) {
		return new Response(JSON.stringify({ error: 'Site is unpublished' }), {
			status: 401,
			headers: { 'content-type': 'application/json' }
		});
	}

	const next = targetWithQuery(event.url);
	const nextParam = encodeURIComponent(next);
	throw redirect(303, `${UNPUBLISHED_ROUTE}?next=${nextParam}`);
};
