import { env } from '$env/dynamic/private';

export const UNPUBLISHED_ACCESS_COOKIE = 'tex_unpublished_access';
const TRUE_VALUES = new Set(['1', 'true', 'yes', 'on']);

function normalize(value: string | undefined): string {
	return value?.trim().toLowerCase() ?? '';
}

export function isUnpublishedModeEnabled(): boolean {
	return TRUE_VALUES.has(normalize(env.UNPUBLISHED_MODE));
}

export function getUnpublishedPassword(): string {
	return env.UNPUBLISHED_PASSWORD?.trim() ?? '';
}

function computeAccessToken(password: string): string {
	const input = `texinroistot:${password}`;
	let hash = 5381;
	for (let i = 0; i < input.length; i += 1) {
		hash = ((hash << 5) + hash) ^ input.charCodeAt(i);
	}
	return `gate_${(hash >>> 0).toString(16)}`;
}

export function expectedUnpublishedAccessToken(): string {
	const password = getUnpublishedPassword();
	if (!password) return '';
	return computeAccessToken(password);
}

export function hasUnpublishedAccess(cookieValue: string | undefined): boolean {
	const expected = expectedUnpublishedAccessToken();
	if (!expected) return false;
	return cookieValue === expected;
}

export function verifyUnpublishedPassword(candidatePassword: string): boolean {
	const password = getUnpublishedPassword();
	if (!password) return false;
	return candidatePassword.trim() === password;
}
