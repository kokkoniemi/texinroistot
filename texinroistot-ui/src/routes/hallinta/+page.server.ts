import { env } from '$env/dynamic/public';
import type { PageServerLoad } from './$types';

type MePayload = {
	loggedIn?: boolean;
	email?: string;
	isAdmin?: boolean;
};

export const load: PageServerLoad = async ({ fetch }) => {
	const googleClientId = env.PUBLIC_GOOGLE_OAUTH2_CLIENT_ID?.trim() ?? '';
	const fallbackData = {
		user: {
			loggedIn: false,
			email: '',
			isAdmin: false
		},
		googleClientId,
		users: [] as { hash: string; isAdmin: boolean; createdAt?: string }[],
		usersError: ''
	};

	try {
		const response = await fetch('/api/me');
		if (!response.ok) {
			return fallbackData;
		}

		const payload = (await response.json()) as MePayload;
		const user = {
			loggedIn: Boolean(payload.loggedIn),
			email: payload.email ?? '',
			isAdmin: Boolean(payload.isAdmin)
		};

		if (!user.loggedIn || !user.isAdmin) {
			return {
				user,
				googleClientId,
				users: [],
				usersError: ''
			};
		}

		const usersResponse = await fetch('/api/admin/users');
		if (!usersResponse.ok) {
			return {
				user,
				googleClientId,
				users: [],
				usersError: 'Käyttäjien haku epäonnistui.'
			};
		}

		const usersPayload = (await usersResponse.json()) as {
			users?: { hash: string; isAdmin: boolean; createdAt?: string }[];
		};
		return {
			user,
			googleClientId,
			users: usersPayload.users ?? [],
			usersError: ''
		};
	} catch {
		return fallbackData;
	}
};
