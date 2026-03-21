import { env } from '$env/dynamic/private';
import type { LayoutServerLoad } from './$types';

type ActiveVersionPayload = {
	version?: {
		createdAt?: string | null;
	};
	stats?: {
		villains?: number | null;
		stories?: number | null;
		drawers?: number | null;
		writers?: number | null;
		translators?: number | null;
	};
};

export const load: LayoutServerLoad = async ({ fetch, url }) => {
	const systemVersion = env.SYSTEM_VERSION?.trim() || 'ei tiedossa';

	if (url.pathname === '/julkaisematon') {
		return {
			systemVersion,
			activeVersionCreatedAt: null,
			activeVersionStats: null
		};
	}

	try {
		const res = await fetch('/api/version/active');
		if (!res.ok) {
			return {
				systemVersion,
				activeVersionCreatedAt: null,
				activeVersionStats: null
			};
		}

		const payload = (await res.json()) as ActiveVersionPayload;
		return {
			systemVersion,
			activeVersionCreatedAt: payload.version?.createdAt ?? null,
			activeVersionStats: payload.stats
				? {
						villains: payload.stats.villains ?? null,
						stories: payload.stats.stories ?? null,
						drawers: payload.stats.drawers ?? null,
						writers: payload.stats.writers ?? null,
						translators: payload.stats.translators ?? null
					}
				: null
		};
	} catch {
		return {
			systemVersion,
			activeVersionCreatedAt: null,
			activeVersionStats: null
		};
	}
};
