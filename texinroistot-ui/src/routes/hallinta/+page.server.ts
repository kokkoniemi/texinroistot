import { env } from '$env/dynamic/public';
import type { PageServerLoad } from './$types';

type MePayload = {
	loggedIn?: boolean;
	email?: string;
};

export const load: PageServerLoad = async ({ fetch }) => {
	const googleClientId = env.PUBLIC_GOOGLE_OAUTH2_CLIENT_ID?.trim() ?? '';

	try {
		const response = await fetch('/api/me');
		if (!response.ok) {
			return {
				user: {
					loggedIn: false,
					email: ''
				},
				googleClientId
			};
		}

		const payload = (await response.json()) as MePayload;
		return {
			user: {
				loggedIn: Boolean(payload.loggedIn),
				email: payload.email ?? ''
			},
			googleClientId
		};
	} catch {
		return {
			user: {
				loggedIn: false,
				email: ''
			},
			googleClientId
		};
	}
};
