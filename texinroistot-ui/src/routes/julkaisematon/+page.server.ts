import { dev } from '$app/environment';
import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import {
	expectedUnpublishedAccessToken,
	getUnpublishedPassword,
	hasUnpublishedAccess,
	isUnpublishedModeEnabled,
	UNPUBLISHED_ACCESS_COOKIE,
	verifyUnpublishedPassword
} from '$lib/server/unpublished-gate';

function safeNext(rawNext: string | null): string {
	if (!rawNext) return '/';
	if (!rawNext.startsWith('/')) return '/';
	if (rawNext.startsWith('//')) return '/';
	if (rawNext.startsWith('/julkaisematon')) return '/';
	return rawNext;
}

export const load: PageServerLoad = async ({ url, cookies }) => {
	if (!isUnpublishedModeEnabled()) {
		throw redirect(303, '/');
	}

	const next = safeNext(url.searchParams.get('next'));
	const hasAccess = hasUnpublishedAccess(cookies.get(UNPUBLISHED_ACCESS_COOKIE));
	if (hasAccess) {
		throw redirect(303, next);
	}

	return {
		next,
		passwordConfigured: getUnpublishedPassword().length > 0
	};
};

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		if (!isUnpublishedModeEnabled()) {
			throw redirect(303, '/');
		}

		const formData = await request.formData();
		const password = String(formData.get('password') ?? '');
		const next = safeNext(String(formData.get('next') ?? '/'));

		if (!getUnpublishedPassword()) {
			return fail(500, {
				next,
				error: 'UNPUBLISHED_PASSWORD puuttuu palvelimen asetuksista.'
			});
		}

		if (!verifyUnpublishedPassword(password)) {
			return fail(400, {
				next,
				error: 'Virheellinen salasana.'
			});
		}

		cookies.set(UNPUBLISHED_ACCESS_COOKIE, expectedUnpublishedAccessToken(), {
			path: '/',
			httpOnly: true,
			sameSite: 'lax',
			secure: !dev,
			maxAge: 60 * 60 * 24 * 14
		});

		throw redirect(303, next);
	}
};
