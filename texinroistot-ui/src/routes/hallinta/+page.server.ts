import { env } from '$env/dynamic/public';
import type { PageServerLoad } from './$types';

type MePayload = {
	loggedIn?: boolean;
	email?: string;
	isAdmin?: boolean;
};

type AdminUser = {
	hash: string;
	isAdmin: boolean;
	createdAt?: string;
};

type AdminVersion = {
	id: number;
	createdAt?: string;
	isActive: boolean;
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
		users: [] as AdminUser[],
		usersError: '',
		versions: [] as AdminVersion[],
		versionsError: '',
		importUrl: ''
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
				usersError: '',
				versions: [],
				versionsError: '',
				importUrl: ''
			};
		}

		let users: AdminUser[] = [];
		let usersError = '';
		const usersResponse = await fetch('/api/admin/users');
		if (usersResponse.ok) {
			const usersPayload = (await usersResponse.json()) as { users?: AdminUser[] };
			users = usersPayload.users ?? [];
		} else {
			usersError = 'Käyttäjien haku epäonnistui.';
		}

		let versions: AdminVersion[] = [];
		let versionsError = '';
		let importUrl = '';
		const versionsResponse = await fetch('/api/admin/versions');
		if (versionsResponse.ok) {
			const versionsPayload = (await versionsResponse.json()) as {
				versions?: AdminVersion[];
				importUrl?: string;
			};
			versions = versionsPayload.versions ?? [];
			importUrl = versionsPayload.importUrl ?? '';
		} else {
			versionsError = 'Versioiden haku epäonnistui.';
		}

		return {
			user,
			googleClientId,
			users,
			usersError,
			versions,
			versionsError,
			importUrl
		};
	} catch {
		return fallbackData;
	}
};
