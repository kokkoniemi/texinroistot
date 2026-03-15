import type { RequestHandler } from './$types';
import { getBackendHost } from '$lib/server/backend-host';
import { authProxyHeaders, proxiedResponse } from '$lib/server/proxy-auth';

export const POST: RequestHandler = async ({ request, fetch }) => {
	const payload = await request.text();
	const headers = authProxyHeaders(request, {
		'content-type': 'application/json'
	});

	const response = await fetch(`${getBackendHost()}/api/admin/users/grant-admin`, {
		method: 'POST',
		headers,
		body: payload
	});

	return proxiedResponse(response);
};
