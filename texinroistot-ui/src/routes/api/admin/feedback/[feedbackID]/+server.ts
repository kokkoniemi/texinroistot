import type { RequestHandler } from './$types';
import { getBackendHost } from '$lib/server/backend-host';
import { authProxyHeaders, proxiedResponse } from '$lib/server/proxy-auth';

export const DELETE: RequestHandler = async ({ request, fetch, params }) => {
	const feedbackID = encodeURIComponent(params.feedbackID);
	const response = await fetch(`${getBackendHost()}/api/admin/feedback/${feedbackID}`, {
		method: 'DELETE',
		headers: authProxyHeaders(request)
	});

	return proxiedResponse(response);
};
