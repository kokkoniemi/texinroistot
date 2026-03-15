import type { RequestHandler } from './$types';
import { getBackendHost } from '$lib/server/backend-host';
import { authProxyHeaders, proxiedResponse } from '$lib/server/proxy-auth';

export const DELETE: RequestHandler = async ({ request, params, fetch }) => {
	const headers = authProxyHeaders(request);
	const versionID = encodeURIComponent(params.versionID);

	const response = await fetch(`${getBackendHost()}/api/admin/versions/${versionID}`, {
		method: 'DELETE',
		headers
	});

	return proxiedResponse(response);
};
