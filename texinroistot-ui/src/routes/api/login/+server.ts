import type { RequestHandler } from './$types';
import { getBackendHost } from '$lib/server/backend-host';
import { authProxyHeaders, proxiedResponse } from '$lib/server/proxy-auth';

export const POST: RequestHandler = async ({ request, fetch }) => {
	const formData = await request.formData();
	const body = new URLSearchParams();
	for (const [key, value] of formData.entries()) {
		if (typeof value === 'string') {
			body.append(key, value);
		}
	}

	const headers = authProxyHeaders(request, {
		'content-type': 'application/x-www-form-urlencoded'
	});

	const response = await fetch(`${getBackendHost()}/api/login`, {
		method: 'POST',
		headers,
		body: body.toString(),
		redirect: 'manual'
	});

	return proxiedResponse(response);
};
