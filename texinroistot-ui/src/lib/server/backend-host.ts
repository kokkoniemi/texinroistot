import { env } from '$env/dynamic/private';

const DEFAULT_BACKEND_HOST = 'http://backend:6969';

export function getBackendHost(): string {
	const configured = env.BACKEND_HOST?.trim();
	return configured && configured.length > 0 ? configured : DEFAULT_BACKEND_HOST;
}
