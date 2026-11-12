import type { RequestHandler } from './$types';
import { getBackendHost } from '$lib/server/backend-host';
import { authProxyHeaders, proxiedResponse } from '$lib/server/proxy-auth';

export const GET: RequestHandler = async ({ request, fetch }) => {
	const response = await fetch(`${getBackendHost()}/api/admin/feedback`, {
		method: 'GET',
		headers: authProxyHeaders(request)
	});

	return proxiedResponse(response);
};
